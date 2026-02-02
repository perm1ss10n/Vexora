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
	"github.com/perm1ss10n/vexora/backend/internal/registry"
)

type Server struct {
	cmd   *commands.Manager
	auth  *auth.Store
	token *auth.TokenService
	reg   *registry.SQLiteStore
}

type SendCmdRequest struct {
	Type      string         `json:"type"`
	Params    map[string]any `json:"params,omitempty"`
	TimeoutMs int            `json:"timeoutMs,omitempty"`
}

type SendCmdResponse struct {
	Ack model.AckPayload `json:"ack"`
}

func New(cmd *commands.Manager, authStore *auth.Store, tokenService *auth.TokenService, registryStore *registry.SQLiteStore) *Server {
	return &Server{cmd: cmd, auth: authStore, token: tokenService, reg: registryStore}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/dev/", s.handleDev)
	if s.auth != nil && s.token != nil {
		authHandler := auth.NewHandler(s.auth, s.token)
		mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
		mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
		mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
		mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)
		mux.Handle("/api/v1/auth/me", auth.RequireAuth(s.token, http.HandlerFunc(authHandler.Me)))
	}
	if s.reg != nil && s.token != nil {
		mux.Handle("/api/v1/devices", auth.RequireAuth(s.token, http.HandlerFunc(s.handleDevices)))
	}
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	return withCORS(mux)
}

type DeviceResponse struct {
	DeviceID        string  `json:"deviceId"`
	Status          string  `json:"status"`
	LastSeen        string  `json:"lastSeen"`
	LastTelemetryAt *string `json:"lastTelemetryAt"`
	FWVersion       *string `json:"fwVersion"`
}

type DevicesResponse struct {
	Devices []DeviceResponse `json:"devices"`
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devices, err := s.reg.ListDevices(r.Context())
	if err != nil {
		log.Printf("[HTTP] list devices failed: %v", err)
		http.Error(w, "failed to list devices", http.StatusInternalServerError)
		return
	}

	response := DevicesResponse{Devices: make([]DeviceResponse, 0, len(devices))}
	for _, device := range devices {
		status := device.Status
		if status == "" {
			status = "offline"
		}

		lastSeen := time.UnixMilli(device.LastSeenMillis).UTC().Format(time.RFC3339)
		var lastTelemetry *string
		if device.TelemetryMillis.Valid {
			value := time.UnixMilli(device.TelemetryMillis.Int64).UTC().Format(time.RFC3339)
			lastTelemetry = &value
		}

		var fwVersion *string
		if device.FW.Valid {
			value := device.FW.String
			if value != "" {
				fwVersion = &value
			}
		}

		response.Devices = append(response.Devices, DeviceResponse{
			DeviceID:        device.DeviceID,
			Status:          status,
			LastSeen:        lastSeen,
			LastTelemetryAt: lastTelemetry,
			FWVersion:       fwVersion,
		})
	}

	writeJSON(w, http.StatusOK, response)
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
