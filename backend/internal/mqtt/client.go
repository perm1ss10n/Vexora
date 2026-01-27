package mqtt

import (
    "fmt"
    "log"
    "os"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
    BrokerURL string
    ClientID  string
    Username  string
    Password  string
}

func LoadConfigFromEnv() Config {
    cfg := Config{
        BrokerURL: getenv("MQTT_BROKER_URL", "tcp://localhost:1883"),
        ClientID:  getenv("MQTT_CLIENT_ID", "vexora-backend-dev"),
        Username:  os.Getenv("MQTT_USERNAME"),
        Password:  os.Getenv("MQTT_PASSWORD"),
    }
    return cfg
}

func NewClient(cfg Config, handler mqtt.MessageHandler) mqtt.Client {
    opts := mqtt.NewClientOptions().
        AddBroker(cfg.BrokerURL).
        SetClientID(cfg.ClientID).
        SetAutoReconnect(true).
        SetConnectRetry(true).
        SetConnectRetryInterval(2 * time.Second)

    if cfg.Username != "" {
        opts.SetUsername(cfg.Username)
        opts.SetPassword(cfg.Password)
    }

    opts.OnConnectionLost = func(_ mqtt.Client, err error) {
        log.Printf("[MQTT] connection lost: %v", err)
    }
    opts.OnConnect = func(c mqtt.Client) {
        log.Printf("[MQTT] connected to %s", cfg.BrokerURL)
        for _, t := range AllTopics {
            token := c.Subscribe(t, 1, handler)
            token.Wait()
            if token.Error() != nil {
                log.Printf("[MQTT] subscribe failed topic=%s err=%v", t, token.Error())
            } else {
                log.Printf("[MQTT] subscribed topic=%s", t)
            }
        }
    }

    return mqtt.NewClient(opts)
}

func Connect(c mqtt.Client) error {
    token := c.Connect()
    token.Wait()
    return token.Error()
}

func getenv(k, def string) string {
    v := os.Getenv(k)
    if v == "" {
        return def
    }
    return v
}

func MustPrintConfig(cfg Config) {
    log.Printf("[MQTT] broker=%s clientId=%s user=%s", cfg.BrokerURL, cfg.ClientID, mask(cfg.Username))
}

func mask(s string) string {
    if s == "" {
        return ""
    }
    return fmt.Sprintf("%c***", s[0])
}