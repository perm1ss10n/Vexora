package main

import (
    "log"

    "github.com/perm1ss10n/vexora/backend/internal/mqtt"
)

func main() {
    handler := mqtt.MakeMessageHandler()

    cfg := mqtt.LoadConfigFromEnv()
    mqtt.MustPrintConfig(cfg)

    c := mqtt.NewClient(cfg, handler)
    if err := mqtt.Connect(c); err != nil {
        log.Fatalf("mqtt connect failed: %v", err)
    }

    // Блокируемся навсегда
    select {}
}