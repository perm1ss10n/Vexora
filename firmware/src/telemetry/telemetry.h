#pragma once

#include "config/config_manager.h"
#include <stdint.h>

class Telemetry
{
public:
    static void init(const TelemetryConfig &cfg, const char *deviceId);
    static void loop();

    // MVP: быстро отправить одну метрику (можно дергать из кода)
    static bool publishMetric(const char *key, float value, const char *unit = nullptr);

    // Runtime updates (3.4.x cfg apply)
    static void updateInterval(uint32_t intervalMs);
    static void updateMinPublishMs(uint32_t minPublishMs);

private:
    static void publishTick();
};