#include "config/config_manager.h"
#include "log/logger.h"

static DeviceConfig g_cfg;

bool ConfigManager::init() {
    g_cfg.valid = true;

    g_cfg.mqtt.host = "broker.hivemq.com"; // временно
    g_cfg.mqtt.port = 1883;
    g_cfg.mqtt.clientId = "vexora-dev";
    g_cfg.mqtt.user = "";
    g_cfg.mqtt.password = "";
    g_cfg.mqtt.lwtTopic = "v1/dev/dev-123/state";
    g_cfg.mqtt.lwtPayloadOnline = "online";
    g_cfg.mqtt.lwtPayloadOffline = "offline";

    g_cfg.telemetry.intervalMs = 5000;

    LOGI("CFG", "config loaded (defaults)");
    return true;
}

DeviceConfig ConfigManager::getActive() {
    return g_cfg;
}