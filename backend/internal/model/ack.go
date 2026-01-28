package model

type AckPayload struct {
	V        int            `json:"v"`
	ID       string         `json:"id"`
	DeviceID string         `json:"deviceId"`
	Ts       int64          `json:"ts"`
	Ok       bool           `json:"ok"`
	Code     string         `json:"code,omitempty"`
	Msg      string         `json:"msg,omitempty"`
	Data     map[string]any `json:"data,omitempty"`
}
