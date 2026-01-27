package mqtt

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/perm1ss10n/vexora/backend/internal/influx"
	"github.com/perm1ss10n/vexora/backend/internal/model"
)

type Dispatcher struct {
	Influx *influx.Client

	mu         sync.Mutex
	lastWrite  map[string]int64 // key = deviceId|metric -> unixMillis
	minWriteMs int64
}

func (d *Dispatcher) InitRateLimitFromEnv() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.lastWrite == nil {
		d.lastWrite = make(map[string]int64)
	}

	d.minWriteMs = 0
	v := os.Getenv("TELEMETRY_MIN_WRITE_MS")
	if v == "" {
		return
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		return
	}
	d.minWriteMs = n
}

func (d *Dispatcher) Dispatch(topic string, payload []byte, env model.Envelope) {
	switch {
	case strings.HasSuffix(topic, "/telemetry"):
		d.handleTelemetry(topic, payload, env)
	case strings.HasSuffix(topic, "/event"):
		d.handleEvent(topic, payload, env)
	case strings.HasSuffix(topic, "/state"):
		d.handleState(topic, payload, env)
	case strings.HasSuffix(topic, "/ack"):
		log.Printf("[ACK] topic=%s deviceId=%s ts=%d size=%d", topic, env.DeviceID, env.Ts, len(payload))
	case strings.HasSuffix(topic, "/cfg/status"):
		log.Printf("[CFG_STATUS] topic=%s deviceId=%s ts=%d size=%d", topic, env.DeviceID, env.Ts, len(payload))
	default:
		log.Printf("[MQTT] topic=%s deviceId=%s ts=%d size=%d", topic, env.DeviceID, env.Ts, len(payload))
	}
}

func (d *Dispatcher) handleTelemetry(topic string, payload []byte, env model.Envelope) {
	var t model.TelemetryPayload
	if err := json.Unmarshal(payload, &t); err != nil {
		log.Printf("[TEL] invalid_json topic=%s err=%v", topic, err)
		return
	}
	if len(t.Metrics) == 0 {
		log.Printf("[TEL] no_metrics topic=%s deviceId=%s ts=%d", topic, env.DeviceID, env.Ts)
		return
	}

	if d.Influx == nil {
		log.Printf("[TEL] influx_disabled topic=%s deviceId=%s metrics=%d", topic, env.DeviceID, len(t.Metrics))
		return
	}

	ts := time.UnixMilli(env.Ts)
	nowMs := time.Now().UnixMilli()

	wrote := 0
	for _, m := range t.Metrics {
		if strings.TrimSpace(m.Key) == "" {
			continue
		}

		// Rate-limit (защита от спайков)
		if d.minWriteMs > 0 {
			key := env.DeviceID + "|" + m.Key

			d.mu.Lock()
			if d.lastWrite == nil {
				d.lastWrite = make(map[string]int64)
			}
			last := d.lastWrite[key]
			if last != 0 && (nowMs-last) < d.minWriteMs {
				d.mu.Unlock()
				continue
			}
			d.lastWrite[key] = nowMs
			d.mu.Unlock()
		}

		p := influxdb2.NewPoint(
			"telemetry",
			map[string]string{
				"deviceId": env.DeviceID,
				"metric":   m.Key,
				"unit":     m.Unit,
			},
			map[string]interface{}{
				"value": m.Value,
			},
			ts,
		)

		d.Influx.WritePoint(p) // async batching
		wrote++
	}
	log.Printf("[TEL] stored topic=%s deviceId=%s metrics=%d wrote=%d", topic, env.DeviceID, len(t.Metrics), wrote)
}

func (d *Dispatcher) handleState(topic string, payload []byte, env model.Envelope) {
	var s model.StatePayload
	if err := json.Unmarshal(payload, &s); err != nil {
		log.Printf("[STATE] invalid_json topic=%s err=%v", topic, err)
		return
	}

	log.Printf("[STATE] recv topic=%s deviceId=%s status=%s ts=%d size=%d", topic, env.DeviceID, s.Status, env.Ts, len(payload))

	if d.Influx == nil {
		return
	}

	ts := time.UnixMilli(env.Ts)

	tags := map[string]string{
		"deviceId": env.DeviceID,
	}
	if s.Status != "" {
		tags["status"] = s.Status
	}
	if s.Link != nil && s.Link.Type != "" {
		tags["link"] = s.Link.Type
	}

	fields := map[string]interface{}{}
	if s.Link != nil {
		if s.Link.Rssi != nil {
			fields["rssi"] = *s.Link.Rssi
		}
		if s.Link.IP != "" {
			fields["ip"] = s.Link.IP
		}
	}
	if s.FW != "" {
		fields["fw"] = s.FW
	}
	if s.UptimeSec != nil {
		fields["uptimeSec"] = *s.UptimeSec
	}
	if s.Cfg != nil {
		if s.Cfg.ActiveVersion != nil {
			fields["cfgActiveVersion"] = *s.Cfg.ActiveVersion
		}
		if s.Cfg.PendingVersion != nil {
			fields["cfgPendingVersion"] = *s.Cfg.PendingVersion
		}
	}

	if len(fields) == 0 {
		// чтобы точка не была пустой
		fields["seen"] = 1
	}

	p := influxdb2.NewPoint("state", tags, fields, ts)
	d.Influx.WritePoint(p)
}

func (d *Dispatcher) handleEvent(topic string, payload []byte, env model.Envelope) {
	var e model.EventPayload
	if err := json.Unmarshal(payload, &e); err != nil {
		log.Printf("[EVENT] invalid_json topic=%s err=%v", topic, err)
		return
	}

	log.Printf("[EVENT] recv topic=%s deviceId=%s code=%s ts=%d size=%d", topic, env.DeviceID, e.Code, env.Ts, len(payload))

	if d.Influx == nil {
		return
	}

	ts := time.UnixMilli(env.Ts)

	tags := map[string]string{
		"deviceId": env.DeviceID,
	}
	if e.Code != "" {
		tags["code"] = e.Code
	}
	if e.Severity != "" {
		tags["severity"] = e.Severity
	}

	fields := map[string]interface{}{}
	if e.Msg != "" {
		fields["msg"] = e.Msg
	}
	if e.Data != nil && len(e.Data) > 0 {
		if b, err := json.Marshal(e.Data); err == nil {
			fields["data"] = string(b)
		}
	}
	if len(fields) == 0 {
		fields["seen"] = 1
	}

	p := influxdb2.NewPoint("event", tags, fields, ts)
	d.Influx.WritePoint(p)
}
