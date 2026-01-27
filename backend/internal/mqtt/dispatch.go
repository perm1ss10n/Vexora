package mqtt

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/perm1ss10n/vexora/backend/internal/influx"
	"github.com/perm1ss10n/vexora/backend/internal/model"
)

type Dispatcher struct {
	Influx *influx.Client
}

func (d *Dispatcher) Dispatch(topic string, payload []byte, env model.Envelope) {
	switch {
	case strings.HasSuffix(topic, "/telemetry"):
		d.handleTelemetry(topic, payload, env)
	case strings.HasSuffix(topic, "/event"):
		log.Printf("[EVENT] topic=%s deviceId=%s ts=%d size=%d", topic, env.DeviceID, env.Ts, len(payload))
	case strings.HasSuffix(topic, "/state"):
		log.Printf("[STATE] topic=%s deviceId=%s ts=%d size=%d", topic, env.DeviceID, env.Ts, len(payload))
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

	wrote := 0
	for _, m := range t.Metrics {
		if strings.TrimSpace(m.Key) == "" {
			continue
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

		if err := d.Influx.WritePoint(p); err != nil {
			log.Printf("[TEL] influx_write_failed deviceId=%s metric=%s err=%v", env.DeviceID, m.Key, err)
			continue
		}
		wrote++
	}
	log.Printf("[TEL] stored topic=%s deviceId=%s metrics=%d wrote=%d", topic, env.DeviceID, len(t.Metrics), wrote)
}
