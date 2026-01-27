package main

import (
	"log"

	"github.com/perm1ss10n/vexora/backend/internal/influx"
	"github.com/perm1ss10n/vexora/backend/internal/mqtt"
)

func main() {
	// Influx (опционально): если токена нет — просто логируем без записи
	var influxClient *influx.Client
	icfg := influx.LoadConfigFromEnv()
	if icfg.Token == "" {
		log.Printf("[INFLUX] disabled (INFLUX_TOKEN is empty)")
	} else {
		influxClient = influx.New(icfg)
		defer influxClient.Close()
		log.Printf("[INFLUX] enabled url=%s org=%s bucket=%s", icfg.URL, icfg.Org, icfg.Bucket)
	}

	d := &mqtt.Dispatcher{Influx: influxClient}
	handler := mqtt.MakeMessageHandler(d)

	cfg := mqtt.LoadConfigFromEnv()
	mqtt.MustPrintConfig(cfg)

	c := mqtt.NewClient(cfg, handler)
	if err := mqtt.Connect(c); err != nil {
		log.Fatalf("mqtt connect failed: %v", err)
	}

	select {}
}
