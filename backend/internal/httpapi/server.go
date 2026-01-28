package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/perm1ss10n/vexora/backend/internal/commands"
	"github.com/perm1ss10n/vexora/backend/internal/model"
)

type Server struct {
	cmd *commands.Manager
}

type SendCmdRequest struct {
	Type      string         `json:"type"`
	Params    map[string]any `json:"params,omitempty"`
	TimeoutMs int            `json:"timeoutMs,omitempty"`
}

type SendCmdResponse struct {
	Ack model.AckPayload `json:"ack"`
}

func New(cmd *commands.Manager) *Server {
	return &Server{cmd: cmd}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/dev/", s.handleDev)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
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
