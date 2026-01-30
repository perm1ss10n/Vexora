#pragma once

#include "config/config_manager.h"

class Telemetry
{
public:
    static void init(const TelemetryConfig &cfg, const char *deviceId);
    static void loop();

    // MVP: быстро отправить одну метрику (можно дергать из кода)
    static bool publishMetric(const char *key, float value, const char *unit = nullptr);
    static void updateInterval(uint32_t intervalMs);

private:
    static void publishTick();
};