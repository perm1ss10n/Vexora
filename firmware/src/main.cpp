#include <Arduino.h>

#include "log/logger.h"
#include "config/config_manager.h"
#include "state/state_machine.h"
#include "mqtt/mqtt_client.h"
#include "telemetry/telemetry.h"
#include "app/app.h"

void setup()
{
    // ===== 1. Hardware / Serial =====
    Serial.begin(115200);
    delay(100);

    // ===== 2. Logger =====
    Logger::init();
    LOGI("BOOT", "Vexora firmware starting...");

    // ===== 3. Load configuration =====
    if (!ConfigManager::init()) {
        LOGE("BOOT", "ConfigManager init failed");
    }

    auto cfg = ConfigManager::getActive();
    if (!cfg.valid) {
        LOGW("BOOT", "No valid config found, using defaults");
    }

    // ===== 4. State machine =====
    StateMachine::init();
    StateMachine::set(State::BOOT);

    // ===== 5. MQTT =====
    if (!MqttClient::init(cfg.mqtt)) {
        LOGE("BOOT", "MQTT init failed");
        StateMachine::set(State::ERROR);
    }

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
    MqttClient::loop();
    Telemetry::loop();
    App::loop();

    StateMachine::loop();

    delay(10);
}