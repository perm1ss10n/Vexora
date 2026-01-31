package main

import (
	"log"
	"net/http"
	"os"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/perm1ss10n/vexora/backend/internal/auth"
	"github.com/perm1ss10n/vexora/backend/internal/commands"
	"github.com/perm1ss10n/vexora/backend/internal/httpapi"
	"github.com/perm1ss10n/vexora/backend/internal/influx"
	"github.com/perm1ss10n/vexora/backend/internal/mqtt"
	"github.com/perm1ss10n/vexora/backend/internal/registry"
)

type pahoPublisher struct {
	c paho.Client
}

func (p pahoPublisher) Publish(topic string, qos byte, retained bool, payload []byte) error {
	tok := p.c.Publish(topic, qos, retained, payload)
	_ = tok.WaitTimeout(5 * time.Second)
	return tok.Error()
}

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

	tokenService, err := auth.NewTokenServiceFromEnv()
	if err != nil {
		log.Fatalf("auth token init failed: %v", err)
	}
	authStore := auth.NewStore(reg.DB())

	d := &mqtt.Dispatcher{
		Influx:   influxClient,
		Registry: reg,
	}
	d.InitRateLimitFromEnv()
	handler := mqtt.MakeMessageHandler(d)

	cfg := mqtt.LoadConfigFromEnv()
	mqtt.MustPrintConfig(cfg)
	lost := make(chan error, 1)

	c := mqtt.NewClient(cfg, handler, lost)

	// Commands + HTTP API (stage 2.4)
	// `c` is already a paho.Client (same type), no assertion needed.
	cmdMgr := commands.New(pahoPublisher{c: c})
	d.Commands = cmdMgr

	if err := mqtt.Connect(c, cfg, lost); err != nil {
		log.Fatalf("mqtt connect failed: %v", err)
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	api := httpapi.New(cmdMgr, authStore, tokenService)
	go func() {
		log.Printf("[HTTP] listening addr=%s", addr)
		if err := http.ListenAndServe(addr, api.Handler()); err != nil {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	log.Printf("[BOOT] backend started (mqtt+registry%s)", func() string {
		if influxClient == nil {
			return ""
		}
		return ",influx"
	}())

	// Блокируемся навсегда
	select {}
}
