#include <Arduino.h>
#include <WiFiClient.h>
#include <ESP.h>

#include "log/logger.h"
#include "config/config_manager.h"
#include "state/state_machine.h"
#include "state/state_publisher.h"
#include "mqtt/mqtt_client.h"
#include "telemetry/telemetry.h"
#include "app/app.h"
#include "link/link_manager.h"
#include "link/wifi_link.h"
#include "link/gsm_link.h"
#include "offline/offline_queue.h"

// dev-<12hex>
static void buildDeviceId(char *out, size_t outSize)
{
    const uint64_t mac = ESP.getEfuseMac();
    const unsigned long long id = (unsigned long long)(mac & 0xFFFFFFFFFFFFULL);
    snprintf(out, outSize, "dev-%012llX", id);
}

// ===== Link layer =====
// TODO(3.3): TEMP — Wi-Fi creds hardcoded for bring-up. Replace with provisioning/NVS.
WifiConfig wifiCfg("YOUR_WIFI_SSID", "YOUR_WIFI_PASSWORD", 15000);

WifiLink wifiLink(wifiCfg);
GsmConfig gsmCfg{};
GsmLink gsmLink(gsmCfg);

LinkManager linkManager(&wifiLink, &gsmLink);
static WiFiClient wifiNetClient;

static void onMqttMessage(const char *topic, const uint8_t *payload, size_t len)
{
    (void)payload;
    char buf[96];
    snprintf(buf, sizeof(buf), "recv topic=%s len=%u", topic ? topic : "(null)", (unsigned)len);
    LOGI("MQTT", buf);
}

void setup()
{
    Serial.begin(115200);
    delay(100);

    Logger::init();
    LOGI("BOOT", "Vexora firmware starting...");

    if (!ConfigManager::init())
    {
        LOGE("BOOT", "ConfigManager init failed");
    }

    auto cfg = ConfigManager::getActive();
    if (!cfg.valid)
    {
        LOGW("BOOT", "No valid config found, using defaults");
    }

    static char deviceId[24];
    buildDeviceId(deviceId, sizeof(deviceId));
    LOGI("BOOT", deviceId);
    OfflineQueue::init(20);

    // Пробрасываем deviceId в MQTT
    cfg.mqtt.deviceId = deviceId;
    cfg.mqtt.clientId = deviceId;

    // Telemetry
    TelemetryConfig tcfg;
    tcfg.intervalMs = cfg.telemetry.intervalMs;
    tcfg.minPublishMs = cfg.telemetry.minPublishMs;
    Telemetry::init(tcfg, deviceId);

    // State publisher
    StatePublishConfig scfg;
    scfg.intervalMs = 5000;
    StatePublisher::init(scfg, deviceId, "fw-0.1.0");

    StateMachine::init();
    StateMachine::set(State::BOOT);

    if (!linkManager.begin())
    {
        LOGW("LINK", "no link available at boot");
        StateMachine::set(State::OFFLINE);
    }

    if (!MqttClient::init(cfg.mqtt, wifiNetClient))
    {
        LOGE("BOOT", "MQTT init failed");
        StateMachine::set(State::ERROR);
    }
    MqttClient::setMessageCallback(onMqttMessage);

    App::init();

    StateMachine::set(State::INIT);
    LOGI("BOOT", "Initialization complete");
}

void loop()
{
    linkManager.loop();

    MqttClient::loop();
    OfflineQueue::flush();
    StatePublisher::loop();
    Telemetry::loop();
    App::loop();

    StateMachine::loop();

    delay(10);
}