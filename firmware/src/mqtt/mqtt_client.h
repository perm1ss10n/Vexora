#pragma once

#include <Arduino.h>
#include <Client.h>
#include <stddef.h>
#include <stdint.h>

#include "mqtt/mqtt_topics.h"

struct MqttConfig {
    // Broker
    const char* host = nullptr;
    uint16_t    port = 1883;

    // Identity
    const char* deviceId = nullptr;

    // Auth
    const char* clientId = nullptr;
    const char* user = nullptr;
    const char* password = nullptr;

    // LWT
    // Если lwtTopic == nullptr → автогенерация v1/dev/<deviceId>/lwt
    const char* lwtTopic = nullptr;

    // Payloads
    const char* lwtPayloadOnline = "online";
    const char* lwtPayloadOffline = "offline";
};

class MqttClient {
public:
    using MessageCallback = void (*)(const char* topic, const uint8_t* payload, size_t len);

    static bool init(const MqttConfig& cfg, Client& netClient);

    static void loop();

    static bool publish(const char* topic, const char* payload, bool retained = false);
    static bool isConnected();

    static void setMessageCallback(MessageCallback cb);

    // безопасно: обновляет внутреннюю копию конфига (без const_cast)
    static void setDeviceId(const char* deviceId);

    static uint32_t reconnectCount();
};