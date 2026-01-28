#pragma once

#include <Arduino.h>
#include <Client.h>
#include <stddef.h>
#include <stdint.h>

#include "mqtt/mqtt_topics.h"

// Конфиг MQTT клиента (минимальный)
struct MqttConfig
{
    // Broker
    const char *host = nullptr;
    uint16_t port = 1883;

    // Identity
    const char *deviceId = nullptr;

    // Auth
    const char *clientId = nullptr;
    const char *user = nullptr;
    const char *password = nullptr;

    // LWT
    // Если lwtTopic == nullptr → автогенерация v1/dev/<deviceId>/lwt
    const char *lwtTopic = nullptr;

    // Payloads
    const char *lwtPayloadOnline = "online";
    const char *lwtPayloadOffline = "offline";
};

class MqttClient
{
public:
    // Callback for incoming MQTT messages (topic + raw payload bytes)
    using MessageCallback = void (*)(const char* topic, const uint8_t* payload, size_t len);

    // Optional: set a callback to receive subscribed messages
    static void setMessageCallback(MessageCallback cb);

    // Optional: update deviceId at runtime (useful after provisioning)
    static void setDeviceId(const char* deviceId);

    // Важно: PubSubClient требует конкретный Arduino Client (WiFiClient/TinyGsmClient/etc)
    static bool init(const MqttConfig &cfg, Client &netClient);
    static void loop();

    static bool publish(const char *topic, const char *payload, bool retained = false);
    static bool isConnected();

    // На будущее (для диагностики/метрик)
    static uint32_t reconnectCount();
};