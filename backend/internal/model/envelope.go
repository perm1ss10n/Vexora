package model

type Envelope struct {
    V        int    `json:"v"`
    DeviceID string `json:"deviceId"`
    Ts       int64  `json:"ts"`
}