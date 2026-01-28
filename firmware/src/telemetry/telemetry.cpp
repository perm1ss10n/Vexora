#include "telemetry/telemetry.h"
#include "log/logger.h"

static TelemetryConfig g_telCfg;
static uint32_t g_last = 0;

void Telemetry::init(const TelemetryConfig& cfg) {
    g_telCfg = cfg;
    g_last = 0;
    LOGI("TEL", "init()");
}

void Telemetry::loop() {
    if (g_telCfg.intervalMs == 0) return;

    uint32_t now = millis();
    if (g_last == 0 || (now - g_last) >= g_telCfg.intervalMs) {
        g_last = now;
        // 3.1: только лог, реальные метрики в 3.3
        LOGI("TEL", "tick");
    }
}