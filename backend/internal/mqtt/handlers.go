package mqtt

import (
    "encoding/json"
    "log"

    mqtt "github.com/eclipse/paho.mqtt.golang"

    "github.com/perm1ss10n/vexora/backend/internal/model"
    "github.com/perm1ss10n/vexora/backend/internal/validate"
)

func MakeMessageHandler() mqtt.MessageHandler {
    return func(_ mqtt.Client, msg mqtt.Message) {
        topic := msg.Topic()
        payload := msg.Payload()

        // Базовая попытка распарсить envelope
        var env model.Envelope
        if err := json.Unmarshal(payload, &env); err != nil {
            log.Printf("[MQTT] topic=%s invalid_json err=%v payload=%q", topic, err, truncate(payload, 512))
            return
        }
        if err := validate.EnvelopeBasic(env); err != nil {
            log.Printf("[MQTT] topic=%s invalid_envelope err=%v deviceId=%q ts=%d v=%d", topic, err, env.DeviceID, env.Ts, env.V)
            return
        }

        log.Printf("[MQTT] topic=%s deviceId=%s ts=%d v=%d size=%d", topic, env.DeviceID, env.Ts, env.V, len(payload))
    }
}

func truncate(b []byte, max int) string {
    if len(b) <= max {
        return string(b)
    }
    return string(b[:max]) + "...(truncated)"
}