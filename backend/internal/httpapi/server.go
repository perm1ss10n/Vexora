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
	"github.com/perm1ss10n/vexora/backend/internal/influx"
	"github.com/perm1ss10n/vexora/backend/internal/model"
	"github.com/perm1ss10n/vexora/backend/internal/registry"
)

type Server struct {
	cmd    *commands.Manager
	auth   *auth.Store
	token  *auth.TokenService
	reg    *registry.SQLiteStore
	influx *influx.Client
}

type SendCmdRequest struct {
	Type      string         `json:"type"`
	Params    map[string]any `json:"params,omitempty"`
	TimeoutMs int            `json:"timeoutMs,omitempty"`
}

type SendCmdResponse struct {
	Ack model.AckPayload `json:"ack"`
}

func New(cmd *commands.Manager, authStore *auth.Store, tokenService *auth.TokenService, registryStore *registry.SQLiteStore, influxClient *influx.Client) *Server {
	return &Server{cmd: cmd, auth: authStore, token: tokenService, reg: registryStore, influx: influxClient}
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
		mux.Handle("/api/v1/devices/", auth.RequireAuth(s.token, http.HandlerFunc(s.handleDeviceDetail)))
	}
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	return withCORS(mux)
}

type DeviceResponse struct {
	DeviceID  string  `json:"deviceId"`
	Status    string  `json:"status"`
	LastSeen  int64   `json:"lastSeen"`
	FWVersion *string `json:"fwVersion"`
}

type DeviceStateResponse struct {
	Uptime int64  `json:"uptime"`
	Link   string `json:"link"`
	IP     string `json:"ip"`
}

type LastTelemetryResponse struct {
	Ts      int64              `json:"ts"`
	Metrics map[string]float64 `json:"metrics"`
}

type DeviceDetailResponse struct {
	Device        DeviceResponse         `json:"device"`
	State         *DeviceStateResponse   `json:"state"`
	LastTelemetry *LastTelemetryResponse `json:"lastTelemetry"`
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

	response := make([]DeviceResponse, 0, len(devices))
	for _, device := range devices {
		status := device.Status
		if status == "" {
			status = "offline"
		}

		lastSeen := time.UnixMilli(device.LastSeenMillis).Unix()

		var fwVersion *string
		if device.FW.Valid {
			value := device.FW.String
			if value != "" {
				fwVersion = &value
			}
		}

		response = append(response, DeviceResponse{
			DeviceID:  device.DeviceID,
			Status:    status,
			LastSeen:  lastSeen,
			FWVersion: fwVersion,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleDeviceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceID := strings.TrimPrefix(r.URL.Path, "/api/v1/devices/")
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" || strings.Contains(deviceID, "/") {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}

	device, err := s.reg.GetDevice(r.Context(), deviceID)
	if err != nil {
		log.Printf("[HTTP] get device failed: %v", err)
		http.Error(w, "failed to get device", http.StatusInternalServerError)
		return
	}
	if device == nil {
		http.Error(w, "device not found", http.StatusNotFound)
		return
	}

	status := device.Status
	if status == "" {
		status = "offline"
	}

	lastSeen := time.UnixMilli(device.LastSeenMillis).Unix()

	var fwVersion *string
	if device.FW.Valid {
		value := device.FW.String
		if value != "" {
			fwVersion = &value
		}
	}

	var state *DeviceStateResponse
	if s.influx != nil {
		lastState, err := s.influx.GetLastState(r.Context(), deviceID)
		if err != nil {
			log.Printf("[HTTP] get device state failed: %v", err)
		} else if lastState != nil {
			state = &DeviceStateResponse{
				Uptime: lastState.Uptime,
				Link:   lastState.Link,
				IP:     lastState.IP,
			}
		}
	}

	var lastTelemetry *LastTelemetryResponse
	if s.influx != nil {
		telemetry, err := s.influx.GetLastTelemetry(r.Context(), deviceID)
		if err != nil {
			log.Printf("[HTTP] get device telemetry failed: %v", err)
		} else if telemetry != nil {
			lastTelemetry = &LastTelemetryResponse{
				Ts:      telemetry.Ts,
				Metrics: telemetry.Metrics,
			}
		}
	}

	writeJSON(w, http.StatusOK, DeviceDetailResponse{
		Device: DeviceResponse{
			DeviceID:  device.DeviceID,
			Status:    status,
			LastSeen:  lastSeen,
			FWVersion: fwVersion,
		},
		State:         state,
		LastTelemetry: lastTelemetry,
	})
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
