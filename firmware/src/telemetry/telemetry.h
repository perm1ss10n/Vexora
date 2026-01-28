#pragma once
#include <Arduino.h>
#include "config/config_manager.h"

class Telemetry {
public:
    static void init(const TelemetryConfig& cfg);
    static void loop();
};