#include "mqtt/mqtt_client.h"
#include "log/logger.h"

static bool connected = false;

bool MqttClient::init(const MqttConfig& cfg) {
    LOGI("MQTT", "init()");
    // Реальная реализация будет в 3.2
    connected = false;
    return true;
}

void MqttClient::loop() {
    // Пока пусто — позже сюда пойдёт клиент
}

bool MqttClient::publish(const char* topic, const char* payload, bool retained) {
    if (!connected) {
        return false;
    }
    return true;
}

bool MqttClient::isConnected() {
    return connected;
}