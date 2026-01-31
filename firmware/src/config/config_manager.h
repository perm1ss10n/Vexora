#pragma once

#include <Arduino.h>
#include <stdint.h>

#include <ArduinoJson.h>

#include "mqtt/mqtt_client.h"
#include "mqtt/mqtt_topics.h"

// ===== Device config =====

struct TelemetryConfig
{
    // How often the device publishes telemetry (cadence)
    uint32_t intervalMs = 5000;

    // Optional per-metric rate limit (0 = disabled)
    uint32_t minPublishMs = 0;
};

struct DeviceConfig
{
    bool valid = false;
    MqttConfig mqtt;
    TelemetryConfig telemetry;
};

// ===== Apply result =====

struct ApplyCfgResult
{
    bool ok = false;
    const char *code = "ERR";
    const char *msg = "";
};

// ===== ConfigManager =====

class ConfigManager
{
public:
    static bool init();

    // Важно: чтобы cfg/status и топики строились корректно
    static void setDeviceId(const char *deviceId);

    static DeviceConfig getActive();

    // apply cfg (partial cfg allowed)
    // ожидаем JSON вида: {"telemetry":{"intervalMs":1234, "minPublishMs":0}}
    static ApplyCfgResult applyCandidate(JsonObject cfg);

    // Должен крутиться в loop()
    // Делает commit pending (если health ok) или rollback
    static void loop();

private:
    static bool validateCandidate(const DeviceConfig &base, JsonObject cfg, DeviceConfig &outCandidate, ApplyCfgResult &outRes);

    static void applyRuntime(const DeviceConfig &cfg);

    static void publishCfgStatus(const char *status);

    static void startPendingWindow();
    static void commitPending();
    static void rollbackPending();

private:
    static DeviceConfig g_active;
    static DeviceConfig g_previous;
    static DeviceConfig g_pending;
    static bool g_hasPending;

    static const char *g_deviceId;

    static uint32_t g_pendingDeadlineMs;
    static bool g_pendingSawMqtt;

    static const uint32_t PENDING_GRACE_MS = 15000; // 15 секунд MVP
};