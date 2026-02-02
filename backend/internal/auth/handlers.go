package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	store *Store
	token *TokenService
}

func NewHandler(store *Store, token *TokenService) *Handler {
	return &Handler{store: store, token: token}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type authResponse struct {
	User        userResponse `json:"user"`
	AccessToken string       `json:"accessToken"`
}

type refreshResponse struct {
	AccessToken string `json:"accessToken"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if err := validateEmail(email); err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, "password too short", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "hash failed", http.StatusInternalServerError)
		return
	}
	now := time.Now()
	user, err := h.store.CreateUser(r.Context(), email, string(hash), now)
	if err != nil {
		if isUniqueConstraint(err) {
			http.Error(w, "email already registered", http.StatusConflict)
			return
		}
		http.Error(w, "create user failed", http.StatusInternalServerError)
		return
	}
	accessToken, sessionToken, err := h.issueSession(r, user.ID)
	if err != nil {
		http.Error(w, "session failed", http.StatusInternalServerError)
		return
	}
	writeRefreshCookie(w, sessionToken)
	writeJSON(w, http.StatusCreated, authResponse{
		User:        userResponse{ID: user.ID, Email: user.Email},
		AccessToken: accessToken,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if err := validateEmail(email); err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	user, passwordHash, err := h.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	accessToken, sessionToken, err := h.issueSession(r, user.ID)
	if err != nil {
		http.Error(w, "session failed", http.StatusInternalServerError)
		return
	}
	writeRefreshCookie(w, sessionToken)
	writeJSON(w, http.StatusOK, authResponse{
		User:        userResponse{ID: user.ID, Email: user.Email},
		AccessToken: accessToken,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	refreshCookie, err := r.Cookie("refreshToken")
	if err != nil || refreshCookie.Value == "" {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}
	now := time.Now()
	refreshHash := hashRefreshToken(refreshCookie.Value)
	session, err := h.store.GetSessionByRefreshHash(r.Context(), refreshHash)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	if session.RevokedAt.Valid || session.ExpiresAt <= now.UnixMilli() {
		http.Error(w, "refresh token expired", http.StatusUnauthorized)
		return
	}
	_ = h.store.RevokeSession(r.Context(), session.ID, now)
	accessToken, newRefresh, err := h.issueSession(r, session.UserID)
	if err != nil {
		http.Error(w, "refresh failed", http.StatusInternalServerError)
		return
	}
	writeRefreshCookie(w, newRefresh)
	writeJSON(w, http.StatusOK, refreshResponse{AccessToken: accessToken})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	refreshCookie, err := r.Cookie("refreshToken")
	if err == nil && refreshCookie.Value != "" {
		refreshHash := hashRefreshToken(refreshCookie.Value)
		session, err := h.store.GetSessionByRefreshHash(r.Context(), refreshHash)
		if err == nil {
			_ = h.store.RevokeSession(r.Context(), session.ID, time.Now())
		}
	}
	clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]userResponse{
		"user": {ID: user.ID, Email: user.Email},
	})
}

func (h *Handler) issueSession(r *http.Request, userID string) (string, string, error) {
	accessToken, err := h.token.GenerateAccessToken(userID)
	if err != nil {
		return "", "", err
	}
	refreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}
	now := time.Now()
	expiresAt := now.Add(RefreshTokenTTL)
	userAgent := r.UserAgent()
	ip := parseIP(r.RemoteAddr)
	if _, err := h.store.CreateSession(r.Context(), userID, refreshHash, expiresAt, userAgent, ip, now); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func generateRefreshToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	return token, hashRefreshToken(token), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writeRefreshCookie(w http.ResponseWriter, token string) {
	secure := os.Getenv("AUTH_COOKIE_SECURE") == "true"
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    token,
		Path:     "/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  time.Now().Add(RefreshTokenTTL),
	})
}

func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("empty email")
	}
	_, err := mail.ParseAddress(email)
	return err
}

func parseIP(addr string) string {
	if addr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE")
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
