package main

import (
	"log"

	"github.com/perm1ss10n/vexora/backend/internal/influx"
	"github.com/perm1ss10n/vexora/backend/internal/mqtt"
	"github.com/perm1ss10n/vexora/backend/internal/registry"
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

	// Registry (SQLite)
	rcfg := registry.LoadSQLiteConfigFromEnv()
	reg, err := registry.NewSQLite(rcfg)
	if err != nil {
		log.Fatalf("registry init failed: %v", err)
	}
	defer reg.Close()
	log.Printf("[REGISTRY] enabled db=%s", rcfg.Path)

	d := &mqtt.Dispatcher{
		Influx:   influxClient,
		Registry: reg,
	}
	d.InitRateLimitFromEnv()
	handler := mqtt.MakeMessageHandler(d)

	cfg := mqtt.LoadConfigFromEnv()
	mqtt.MustPrintConfig(cfg)

	c := mqtt.NewClient(cfg, handler)
	if err := mqtt.Connect(c); err != nil {
		log.Fatalf("mqtt connect failed: %v", err)
	}

	select {}
}
