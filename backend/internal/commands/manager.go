package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/perm1ss10n/vexora/backend/internal/model"
)

var (
	ErrTimeout      = errors.New("ack timeout")
	ErrPublish      = errors.New("publish failed")
	ErrInvalidAck   = errors.New("invalid ack")
	ErrNotConnected = errors.New("mqtt not connected")
)

// Publisher — абстракция публикации в MQTT (чтобы не тащить paho сюда напрямую).
type Publisher interface {
	Publish(topic string, qos byte, retained bool, payload []byte) error
}

type Manager struct {
	pub Publisher

	mu      sync.Mutex
	pending map[string]chan model.AckPayload // key = cmdId
}

func New(pub Publisher) *Manager {
	return &Manager{
		pub:     pub,
		pending: make(map[string]chan model.AckPayload),
	}
}

func (m *Manager) Send(ctx context.Context, deviceID string, cmdType string, params map[string]any, timeout time.Duration) (model.AckPayload, error) {
	if deviceID == "" || cmdType == "" {
		return model.AckPayload{}, fmt.Errorf("deviceId/cmdType is empty")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	cmdID := uuid.NewString()
	now := time.Now().UnixMilli()

	cmd := model.CommandPayload{
		V:        1,
		ID:       cmdID,
		DeviceID: deviceID,
		Ts:       now,
		Type:     cmdType,
		Params:   params,
	}

	b, err := json.Marshal(cmd)
	if err != nil {
		return model.AckPayload{}, fmt.Errorf("marshal cmd: %w", err)
	}

	ch := make(chan model.AckPayload, 1)

	// register pending
	m.mu.Lock()
	m.pending[cmdID] = ch
	m.mu.Unlock()

	// cleanup on exit
	defer func() {
		m.mu.Lock()
		delete(m.pending, cmdID)
		m.mu.Unlock()
	}()

	// publish
	topic := fmt.Sprintf("v1/dev/%s/cmd", deviceID)
	if err := m.pub.Publish(topic, 1, false, b); err != nil {
		return model.AckPayload{}, fmt.Errorf("%w: %v", ErrPublish, err)
	}

	// wait ack
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case ack := <-ch:
		if ack.ID == "" || ack.DeviceID == "" {
			return model.AckPayload{}, ErrInvalidAck
		}
		return ack, nil
	case <-timer.C:
		return model.AckPayload{
			V:        1,
			ID:       cmdID,
			DeviceID: deviceID,
			Ts:       time.Now().UnixMilli(),
			Ok:       false,
			Code:     "TIMEOUT",
			Msg:      "ACK timeout",
		}, ErrTimeout
	case <-ctx.Done():
		return model.AckPayload{}, ctx.Err()
	}
}

// OnAck дергается из MQTT dispatcher при получении v1/dev/{deviceId}/ack
func (m *Manager) OnAck(ack model.AckPayload) {
	if ack.ID == "" {
		return
	}
	m.mu.Lock()
	ch := m.pending[ack.ID]
	m.mu.Unlock()

	if ch == nil {
		return
	}

	// не блокируемся
	select {
	case ch <- ack:
	default:
	}
}
