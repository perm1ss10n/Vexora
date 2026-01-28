package model

type CommandPayload struct {
	V        int            `json:"v"`
	ID       string         `json:"id"`
	DeviceID string         `json:"deviceId"`
	Ts       int64          `json:"ts"`
	Type     string         `json:"type"`
	Params   map[string]any `json:"params,omitempty"`
}
