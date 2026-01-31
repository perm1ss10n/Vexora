package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/perm1ss10n/vexora/backend/internal/auth"
	"github.com/perm1ss10n/vexora/backend/internal/commands"
	"github.com/perm1ss10n/vexora/backend/internal/model"
)

type Server struct {
	cmd   *commands.Manager
	auth  *auth.Store
	token *auth.TokenService
}

type SendCmdRequest struct {
	Type      string         `json:"type"`
	Params    map[string]any `json:"params,omitempty"`
	TimeoutMs int            `json:"timeoutMs,omitempty"`
}

type SendCmdResponse struct {
	Ack model.AckPayload `json:"ack"`
}

func New(cmd *commands.Manager, authStore *auth.Store, tokenService *auth.TokenService) *Server {
	return &Server{cmd: cmd, auth: authStore, token: tokenService}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/dev/", s.handleDev)
	if s.auth != nil && s.token != nil {
		authHandler := auth.NewHandler(s.auth, s.token)
		mux.HandleFunc("/v1/auth/register", authHandler.Register)
		mux.HandleFunc("/v1/auth/login", authHandler.Login)
		mux.HandleFunc("/v1/auth/refresh", authHandler.Refresh)
		mux.HandleFunc("/v1/auth/logout", authHandler.Logout)
		mux.Handle("/v1/auth/me", auth.RequireAuth(s.token, http.HandlerFunc(authHandler.Me)))
	}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	return withCORS(mux)
}

func (s *Server) handleDev(w http.ResponseWriter, r *http.Request) {
	// ожидаем: POST /api/v1/dev/{deviceId}/cmd
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/dev/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	deviceID := parts[0]
	action := parts[1]
	if deviceID == "" || action != "cmd" {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}

	var req SendCmdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	timeout := 10 * time.Second
	if req.TimeoutMs > 0 {
		timeout = time.Duration(req.TimeoutMs) * time.Millisecond
	}

	ack, err := s.cmd.Send(r.Context(), deviceID, req.Type, req.Params, timeout)
	if err != nil {
		// TIMEOUT — это не “500”, это ожидаемое поведение
		if strings.Contains(err.Error(), "timeout") {
			writeJSON(w, http.StatusGatewayTimeout, SendCmdResponse{Ack: ack})
			return
		}
		log.Printf("[HTTP] cmd_send_failed deviceId=%s type=%s err=%v", deviceID, req.Type, err)
		writeJSON(w, http.StatusInternalServerError, SendCmdResponse{Ack: ack})
		return
	}

	writeJSON(w, http.StatusOK, SendCmdResponse{Ack: ack})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigin := os.Getenv("WEB_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
