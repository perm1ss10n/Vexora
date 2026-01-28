#include <Arduino.h>
#include <WiFiClient.h>

#include "log/logger.h"
#include "config/config_manager.h"
#include "state/state_machine.h"
#include "mqtt/mqtt_client.h"
#include "telemetry/telemetry.h"
#include "app/app.h"
#include "link/link_manager.h"
#include "link/wifi_link.h"
#include "link/gsm_link.h"

// ===== Link layer =====
// TODO(3.3): move Wi-Fi creds to ConfigManager / provisioning
// TODO(3.3): TEMP — Wi-Fi creds are hardcoded for bring-up.
// Replace with provisioning/app-config stored in NVS.
WifiConfig wifiCfg("YOUR_WIFI_SSID", "YOUR_WIFI_PASSWORD", 15000);

WifiLink wifiLink(wifiCfg);
GsmConfig gsmCfg{};
GsmLink gsmLink(gsmCfg);

LinkManager linkManager(&wifiLink, &gsmLink);


static WiFiClient wifiNetClient;

// ===== MQTT callback (3.2) =====
static void onMqttMessage(const char* topic, const uint8_t* payload, size_t len)
{
    // Пока только логируем входящие. В 3.3/3.4 тут будет роутинг cmd/cfg/ota.
    (void)payload;
    char buf[96];
    snprintf(buf, sizeof(buf), "recv topic=%s len=%u", topic ? topic : "(null)", (unsigned)len);
    LOGI("MQTT", buf);
}

void setup()
{
    // ===== 1. Hardware / Serial =====
    Serial.begin(115200);
    delay(100);

    // ===== 2. Logger =====
    Logger::init();
    LOGI("BOOT", "Vexora firmware starting...");

    // ===== 3. Load configuration =====
    if (!ConfigManager::init())
    {
        LOGE("BOOT", "ConfigManager init failed");
    }

    auto cfg = ConfigManager::getActive();

    if (!cfg.valid)
    {
        LOGW("BOOT", "No valid config found, using defaults");
    }
    cfg.mqtt.deviceId = "dev-123"; // TODO: real deviceId from config/flash
    cfg.mqtt.clientId = "dev-123";

    // ===== 4. State machine =====
    StateMachine::init();
    StateMachine::set(State::BOOT);
    // ===== Link (Wi-Fi/GSM) =====
    if (!linkManager.begin())
    {
        LOGW("LINK", "no link available at boot");
        StateMachine::set(State::OFFLINE);
    }
    else
    {
        auto st = linkManager.status();
        LOGI("LINK", st.type == LinkType::WIFI ? "wifi active" : "gsm active");
    }

    // ===== 5. MQTT =====
    if (!MqttClient::init(cfg.mqtt, wifiNetClient))
    {
        LOGE("BOOT", "MQTT init failed");
        StateMachine::set(State::ERROR);
    }

    // Receive subscribed messages (cmd/cfg/ota)
    MqttClient::setMessageCallback(onMqttMessage);

    // ===== 6. Telemetry =====
    Telemetry::init(cfg.telemetry);

    // ===== 7. App lifecycle =====
    App::init();

    StateMachine::set(State::INIT);
    LOGI("BOOT", "Initialization complete");
}

void loop()
{
    // ===== Core loops =====
    linkManager.loop();

    MqttClient::loop();
    Telemetry::loop();
    App::loop();

    StateMachine::loop();

    delay(10);
}