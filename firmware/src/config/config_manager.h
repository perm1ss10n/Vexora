#pragma once
#include <Arduino.h>
#include "mqtt/mqtt_client.h"   // ← ВАЖНО

struct TelemetryConfig {
    uint32_t intervalMs = 5000;
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