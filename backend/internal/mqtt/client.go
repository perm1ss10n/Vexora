package mqtt

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"
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

// NewClient creates an MQTT client with:
// - OnConnect: subscribes to all topics in AllTopics (re-subscribes after reconnect)
// - OnConnectionLost: pushes the error to `lost` channel (non-blocking) for reconnect loop
func NewClient(cfg Config, handler mqtt.MessageHandler, lost chan<- error) mqtt.Client {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientID).
		SetAutoReconnect(false). // we manage reconnect ourselves
		SetConnectRetry(false)   // avoid internal retry loops

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}

	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Printf("[MQTT] connection_lost err=%v", err)
		if lost != nil {
			select {
			case lost <- err:
			default:
			}
		}
	}

	opts.OnConnect = func(c mqtt.Client) {
		log.Printf("[MQTT] connected to %s", cfg.BrokerURL)
		for _, t := range AllTopics {
			token := c.Subscribe(t, 1, handler)
			token.Wait()
			if token.Error() != nil {
				log.Printf("[MQTT] subscribe_failed topic=%s err=%v", t, token.Error())
			} else {
				log.Printf("[MQTT] subscribed topic=%s", t)
			}
		}
	}

	return mqtt.NewClient(opts)
}

var reconnectCount uint64

// Connect does the initial connect and then starts a reconnect loop with exponential backoff.
// It returns only if the INITIAL connect fails.
func Connect(c mqtt.Client, cfg Config, lost <-chan error) error {
	tok := c.Connect()
	tok.Wait()
	if err := tok.Error(); err != nil {
		return err
	}

	// Reconnect loop (background)
	go reconnectLoop(c, lost)
	return nil
}

func reconnectLoop(c mqtt.Client, lost <-chan error) {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for range lost {
		for {
			cnt := atomic.AddUint64(&reconnectCount, 1)
			log.Printf("[MQTT] reconnecting reconnect_count=%d retry_in=%s", cnt, backoff)
			time.Sleep(backoff)

			tok := c.Connect()
			tok.Wait()
			if tok.Error() == nil {
				backoff = time.Second // reset after success
				break
			}

			// exponential backoff with cap
			backoff = minDur(backoff*2, maxBackoff)
		}
	}
}

func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
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
