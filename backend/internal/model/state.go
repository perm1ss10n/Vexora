package model

// StatePayload — payload для v1/dev/{deviceId}/state (retained)
type StatePayload struct {
	V         int            `json:"v"`
	DeviceID  string         `json:"deviceId"`
	Ts        int64          `json:"ts"`
	Status    string         `json:"status"`         // online/offline/degraded/error
	Link      *LinkState     `json:"link,omitempty"` // wifi/gsm
	FW        string         `json:"fw,omitempty"`   // firmware version
	UptimeSec *int64         `json:"uptimeSec,omitempty"`
	Cfg       *CfgState      `json:"cfg,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

type LinkState struct {
	Type string   `json:"type"` // wifi/gsm
	Rssi *float64 `json:"rssi,omitempty"`
	IP   string   `json:"ip,omitempty"`
}

type CfgState struct {
	ActiveVersion  *int64 `json:"activeVersion,omitempty"`
	PendingVersion *int64 `json:"pendingVersion,omitempty"`
}
