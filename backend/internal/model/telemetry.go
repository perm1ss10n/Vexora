package model

type TelemetryPayload struct {
	V        int     `json:"v"`
	DeviceID string  `json:"deviceId"`
	Ts       int64   `json:"ts"`
	Seq      *int64  `json:"seq,omitempty"`
	Metrics  []Metric `json:"metrics"`
	Meta     map[string]any `json:"meta,omitempty"`
}

type Metric struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
}
