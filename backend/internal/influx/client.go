package influx

import (
	"log"
	"os"
	"strconv"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Client struct {
	cli      influxdb2.Client
	writeAPI api.WriteAPI
	done     chan struct{}
	org      string
	bucket   string
}

type Config struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

func LoadConfigFromEnv() Config {
	return Config{
		URL:    getenv("INFLUX_URL", "http://localhost:8086"),
		Token:  os.Getenv("INFLUX_TOKEN"),
		Org:    getenv("INFLUX_ORG", "vexora"),
		Bucket: getenv("INFLUX_BUCKET", "telemetry"),
	}
}

func New(cfg Config) *Client {
	opts := influxdb2.DefaultOptions().
		SetBatchSize(getenvUint("INFLUX_BATCH_SIZE", 500)).
		SetFlushInterval(getenvUint("INFLUX_FLUSH_MS", 1000)).
		//Отсутствует в нашей версии похоже, для MVP пока нахуй не упёрлась
		// SetRetryJitter(getenvInt("INFLUX_JITTER_MS", 250)).
		SetRetryInterval(getenvUint("INFLUX_RETRY_MS", 1000)).
		SetMaxRetries(getenvUint("INFLUX_MAX_RETRIES", 3)).
		SetMaxRetryTime(getenvUint("INFLUX_MAX_RETRY_TIME_MS", 30000)).
		SetHTTPRequestTimeout(getenvUint("INFLUX_HTTP_TIMEOUT_MS", 5000)).
		SetHTTPRequestTimeout(getenvUint("INFLUX_HTTP_TIMEOUT_MS", 5000))

	cli := influxdb2.NewClientWithOptions(cfg.URL, cfg.Token, opts)
	w := cli.WriteAPI(cfg.Org, cfg.Bucket)

	c := &Client{
		cli:      cli,
		writeAPI: w,
		done:     make(chan struct{}),
		org:      cfg.Org,
		bucket:   cfg.Bucket,
	}

	// Логируем async ошибки записи
	go func() {
		for {
			select {
			case err, ok := <-w.Errors():
				if !ok {
					return
				}
				log.Printf("[INFLUX] async_write_error: %v", err)
			case <-c.done:
				return
			}
		}
	}()

	return c
}

func (c *Client) Close() {
	if c == nil || c.cli == nil {
		return
	}
	close(c.done)
	c.writeAPI.Flush()
	c.cli.Close()
}

func (c *Client) WritePoint(p *write.Point) {
	c.writeAPI.WritePoint(p)
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvUint(k string, def uint) uint {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return def
	}
	return uint(n)
}
