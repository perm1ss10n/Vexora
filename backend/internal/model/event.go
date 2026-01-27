package model

// EventPayload — payload для v1/dev/{deviceId}/event
type EventPayload struct {
	V        int            `json:"v"`
	DeviceID string         `json:"deviceId"`
	Ts       int64          `json:"ts"`
	Code     string         `json:"code"`               // STATE_ONLINE, OTA_FAIL, ...
	Severity string         `json:"severity,omitempty"` // info/warn/error
	Msg      string         `json:"msg,omitempty"`
	Data     map[string]any `json:"data,omitempty"`
}
