package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type LinkState struct {
	Type string   `json:"type"`
	Rssi *float64 `json:"rssi,omitempty"`
	IP   string   `json:"ip,omitempty"`
}

type CfgState struct {
	ActiveVersion  *int64 `json:"activeVersion,omitempty"`
	PendingVersion *int64 `json:"pendingVersion,omitempty"`
}

type StatePayload struct {
	V         int            `json:"v"`
	DeviceID  string         `json:"deviceId"`
	Ts        int64          `json:"ts"`
	Status    string         `json:"status"`
	Link      *LinkState     `json:"link,omitempty"`
	FW        string         `json:"fw,omitempty"`
	UptimeSec *int64         `json:"uptimeSec,omitempty"`
	Cfg       *CfgState      `json:"cfg,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

type Metric struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
}

type TelemetryPayload struct {
	V        int            `json:"v"`
	DeviceID string         `json:"deviceId"`
	Ts       int64          `json:"ts"`
	Seq      *int64         `json:"seq,omitempty"`
	Metrics  []Metric       `json:"metrics"`
	Meta     map[string]any `json:"meta,omitempty"`
}

type CommandPayload struct {
	V        int            `json:"v"`
	ID       string         `json:"id"`
	DeviceID string         `json:"deviceId"`
	Ts       int64          `json:"ts"`
	Type     string         `json:"type"`
	Params   map[string]any `json:"params,omitempty"`
}

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

type EventPayload struct {
	V        int            `json:"v"`
	DeviceID string         `json:"deviceId"`
	Ts       int64          `json:"ts"`
	Code     string         `json:"code"`
	Severity string         `json:"severity,omitempty"`
	Msg      string         `json:"msg,omitempty"`
	Data     map[string]any `json:"data,omitempty"`
}

type simulator struct {
	client    mqtt.Client
	deviceID  string
	fwVersion string

	startedAt time.Time
	baseTS    int64

	mu             sync.RWMutex
	online         bool
	telemetryEvery time.Duration
	minPublishMs   int64
	seq            int64
	activeCfgVer   int64
	pendingCfgVer  int64

	stateTopic string
	telTopic   string
	cmdTopic   string
	ackTopic   string
	eventTopic string
}

func main() {
	var (
		deviceID   = flag.String("device-id", "dev-123", "Device ID")
		mqttURL    = flag.String("mqtt-url", envOrDefault("MQTT_URL", "tcp://localhost:1883"), "MQTT broker URL")
		intervalMs = flag.Int("interval-ms", 5000, "Telemetry interval in milliseconds")
		fwVersion  = flag.String("fw", "fw-0.1.0", "Firmware version")
		online     = flag.Bool("online", true, "Start device in online status")
	)
	flag.Parse()

	if *deviceID == "" {
		log.Fatal("--device-id is required")
	}
	if *intervalMs < 1000 || *intervalMs > 3600000 {
		log.Fatal("--interval-ms must be between 1000 and 3600000")
	}

	sim := &simulator{
		deviceID:       *deviceID,
		fwVersion:      *fwVersion,
		startedAt:      time.Now(),
		baseTS:         time.Now().UnixMilli(),
		online:         *online,
		telemetryEvery: time.Duration(*intervalMs) * time.Millisecond,
		minPublishMs:   0,
		activeCfgVer:   1,
		stateTopic:     fmt.Sprintf("v1/dev/%s/state", *deviceID),
		telTopic:       fmt.Sprintf("v1/dev/%s/telemetry", *deviceID),
		cmdTopic:       fmt.Sprintf("v1/dev/%s/cmd", *deviceID),
		ackTopic:       fmt.Sprintf("v1/dev/%s/ack", *deviceID),
		eventTopic:     fmt.Sprintf("v1/dev/%s/event", *deviceID),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(*mqttURL)
	opts.SetClientID(fmt.Sprintf("device-sim-%s-%d", *deviceID, time.Now().UnixNano()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("[sim] connected to %s", *mqttURL)
		if token := c.Subscribe(sim.cmdTopic, 1, sim.onCommand); token.Wait() && token.Error() != nil {
			log.Printf("[sim] subscribe failed topic=%s err=%v", sim.cmdTopic, token.Error())
			return
		}
		log.Printf("[sim] subscribed topic=%s", sim.cmdTopic)
		if err := sim.publishState(); err != nil {
			log.Printf("[sim] publish state failed: %v", err)
		}
	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("[sim] connection lost: %v", err)
	})

	sim.client = mqtt.NewClient(opts)
	if token := sim.client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("connect mqtt: %v", token.Error())
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go sim.telemetryLoop()

	<-stop
	log.Printf("[sim] shutdown requested")
	sim.client.Disconnect(500)
}

func (s *simulator) telemetryLoop() {
	for {
		s.mu.RLock()
		interval := s.telemetryEvery
		online := s.online
		s.mu.RUnlock()

		time.Sleep(interval)
		if !online {
			continue
		}
		if err := s.publishTelemetry(); err != nil {
			log.Printf("[sim] publish telemetry failed: %v", err)
		}
	}
}

func (s *simulator) onCommand(_ mqtt.Client, msg mqtt.Message) {
	var cmd CommandPayload
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		log.Printf("[sim] invalid cmd payload err=%v payload=%s", err, string(msg.Payload()))
		return
	}

	if cmd.ID == "" {
		cmd.ID = fmt.Sprintf("sim-%d", time.Now().UnixNano())
	}
	if cmd.DeviceID == "" {
		cmd.DeviceID = s.deviceID
	}
	if cmd.Ts == 0 {
		cmd.Ts = time.Now().UnixMilli()
	}

	log.Printf("[sim] cmd recv type=%s id=%s", cmd.Type, cmd.ID)
	switch strings.ToLower(cmd.Type) {
	case "ping":
		s.publishAck(cmd.ID, true, "OK", "pong", nil)
		s.publishEvent("PING", "info", "ping command processed", map[string]any{"id": cmd.ID})
	case "get_state":
		s.publishState()
		s.publishAck(cmd.ID, true, "OK", "state published", nil)
	case "reboot":
		s.publishAck(cmd.ID, true, "OK", "reboot scheduled", nil)
		s.publishEvent("REBOOT", "warn", "device reboot", map[string]any{"id": cmd.ID})
		go s.simulateReboot()
	case "apply_cfg":
		s.handleApplyCfg(cmd)
	default:
		s.publishAck(cmd.ID, false, "UNSUPPORTED_CMD", "unsupported command", nil)
	}
}

func (s *simulator) handleApplyCfg(cmd CommandPayload) {
	cfgRaw, ok := cmd.Params["cfg"]
	if !ok {
		s.publishAck(cmd.ID, false, "BAD_CFG", "missing cfg", nil)
		return
	}

	cfgMap, ok := cfgRaw.(map[string]any)
	if !ok {
		s.publishAck(cmd.ID, false, "BAD_CFG", "cfg must be object", nil)
		return
	}

	telRaw, ok := cfgMap["telemetry"]
	if !ok {
		s.publishAck(cmd.ID, false, "BAD_CFG", "missing cfg.telemetry", nil)
		return
	}
	telMap, ok := telRaw.(map[string]any)
	if !ok {
		s.publishAck(cmd.ID, false, "BAD_CFG", "cfg.telemetry must be object", nil)
		return
	}

	s.mu.Lock()
	newIntervalMs := int64(s.telemetryEvery / time.Millisecond)
	newMinPublish := s.minPublishMs

	if raw, exists := telMap["intervalMs"]; exists {
		v, err := toInt64(raw)
		if err != nil || v < 1000 || v > 3600000 {
			s.mu.Unlock()
			s.publishAck(cmd.ID, false, "CFG_REJECTED", "intervalMs out of range", nil)
			return
		}
		newIntervalMs = v
	}
	if raw, exists := telMap["minPublishMs"]; exists {
		v, err := toInt64(raw)
		if err != nil || v < 0 || v > 3600000 {
			s.mu.Unlock()
			s.publishAck(cmd.ID, false, "CFG_REJECTED", "minPublishMs out of range", nil)
			return
		}
		newMinPublish = v
	}
	if newMinPublish > newIntervalMs {
		s.mu.Unlock()
		s.publishAck(cmd.ID, false, "CFG_REJECTED", "minPublishMs must be <= intervalMs", nil)
		return
	}

	s.telemetryEvery = time.Duration(newIntervalMs) * time.Millisecond
	s.minPublishMs = newMinPublish
	s.activeCfgVer++
	s.pendingCfgVer = 0
	s.mu.Unlock()

	s.publishState()
	s.publishAck(cmd.ID, true, "OK", "config applied", map[string]any{
		"telemetry": map[string]any{
			"intervalMs":   newIntervalMs,
			"minPublishMs": newMinPublish,
		},
	})
	s.publishEvent("CFG_APPLIED", "info", "config applied", map[string]any{"id": cmd.ID})
}

func (s *simulator) simulateReboot() {
	s.mu.Lock()
	s.online = false
	s.mu.Unlock()
	s.publishState()

	time.Sleep(1500 * time.Millisecond)

	s.mu.Lock()
	s.online = true
	s.startedAt = time.Now()
	s.mu.Unlock()
	s.publishState()
	s.publishEvent("STATE_ONLINE", "info", "device back online", nil)
}

func (s *simulator) publishState() error {
	s.mu.RLock()
	online := s.online
	activeVer := s.activeCfgVer
	pendingVer := s.pendingCfgVer
	s.mu.RUnlock()

	now := time.Now().UnixMilli()
	status := "offline"
	if online {
		status = "online"
	}

	rssi := -62.0 + rand.Float64()*8
	uptimeSec := int64(time.Since(s.startedAt).Seconds())
	payload := StatePayload{
		V:        1,
		DeviceID: s.deviceID,
		Ts:       now,
		Status:   status,
		Link: &LinkState{
			Type: "wifi",
			Rssi: &rssi,
			IP:   "192.168.1.123",
		},
		FW:        s.fwVersion,
		UptimeSec: &uptimeSec,
		Cfg: &CfgState{
			ActiveVersion:  &activeVer,
			PendingVersion: &pendingVer,
		},
		Meta: map[string]any{
			"sim": true,
			"telemetry": map[string]any{
				"intervalMs":   int64(s.telemetryEvery / time.Millisecond),
				"minPublishMs": s.minPublishMs,
			},
		},
	}
	return s.publishJSON(s.stateTopic, 1, true, payload)
}

func (s *simulator) publishTelemetry() error {
	now := time.Now().UnixMilli()
	s.mu.Lock()
	s.seq++
	seq := s.seq
	s.mu.Unlock()

	elapsed := float64(now-s.baseTS) / 1000.0
	temp := 24.0 + math.Sin(elapsed/16.0)*2.4 + rand.Float64()*0.4
	heatFlux := 120.0 + math.Sin(elapsed/12.0)*15 + rand.Float64()*1.5
	rssi := -64.0 + math.Sin(elapsed/10.0)*3 + rand.Float64()*0.6
	batt := 3.95 + math.Sin(elapsed/40.0)*0.03 + rand.Float64()*0.01

	payload := TelemetryPayload{
		V:        1,
		DeviceID: s.deviceID,
		Ts:       now,
		Seq:      &seq,
		Metrics: []Metric{
			{Key: "tempC", Value: temp, Unit: "C"},
			{Key: "heatFlux", Value: heatFlux, Unit: "W/m2"},
			{Key: "rssi", Value: rssi, Unit: "dBm"},
			{Key: "battV", Value: batt, Unit: "V"},
		},
		Meta: map[string]any{"sim": true},
	}
	return s.publishJSON(s.telTopic, 0, false, payload)
}

func (s *simulator) publishAck(id string, ok bool, code, msg string, data map[string]any) {
	ack := AckPayload{
		V:        1,
		ID:       id,
		DeviceID: s.deviceID,
		Ts:       time.Now().UnixMilli(),
		Ok:       ok,
		Code:     code,
		Msg:      msg,
		Data:     data,
	}
	if err := s.publishJSON(s.ackTopic, 1, false, ack); err != nil {
		log.Printf("[sim] publish ack failed: %v", err)
	}
}

func (s *simulator) publishEvent(code, severity, msg string, data map[string]any) {
	e := EventPayload{
		V:        1,
		DeviceID: s.deviceID,
		Ts:       time.Now().UnixMilli(),
		Code:     code,
		Severity: severity,
		Msg:      msg,
		Data:     data,
	}
	if err := s.publishJSON(s.eventTopic, 1, false, e); err != nil {
		log.Printf("[sim] publish event failed: %v", err)
	}
}

func (s *simulator) publishJSON(topic string, qos byte, retained bool, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	token := s.client.Publish(topic, qos, retained, b)
	if !token.WaitTimeout(5 * time.Second) {
		return errors.New("mqtt publish timeout")
	}
	if token.Error() != nil {
		return token.Error()
	}
	log.Printf("[sim] pub topic=%s retained=%v payload=%s", topic, retained, string(b))
	return nil
}

func toInt64(v any) (int64, error) {
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case float32:
		return int64(n), nil
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case int32:
		return int64(n), nil
	case json.Number:
		return n.Int64()
	default:
		return 0, fmt.Errorf("unsupported number type %T", v)
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
