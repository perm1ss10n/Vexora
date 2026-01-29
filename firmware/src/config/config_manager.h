#pragma once
#include <Arduino.h>
#include "mqtt/mqtt_client.h"   // ← ВАЖНО

struct TelemetryConfig {
    // How often the device publishes telemetry (cadence)
    uint32_t intervalMs = 5000;

    // Optional per-metric rate limit to protect backend/Influx from spikes (0 = disabled)
    uint32_t minPublishMs = 0;
};

struct DeviceConfig {
    bool valid = false;
    MqttConfig mqtt;
    TelemetryConfig telemetry;
};

class ConfigManager {
public:
    static bool init();
    static DeviceConfig getActive();
};