package influx

import (
	"context"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Client struct {
	cli      influxdb2.Client
	writeAPI api.WriteAPIBlocking
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
	cli := influxdb2.NewClient(cfg.URL, cfg.Token)
	return &Client{
		cli:      cli,
		writeAPI: cli.WriteAPIBlocking(cfg.Org, cfg.Bucket),
	}
}

func (c *Client) Close() {
	if c != nil && c.cli != nil {
		c.cli.Close()
	}
}

func (c *Client) WritePoint(p *write.Point) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.writeAPI.WritePoint(ctx, p)
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
