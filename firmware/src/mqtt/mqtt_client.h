#pragma once

#include <Arduino.h>
#include "mqtt/mqtt_topics.h"

struct MqttConfig {
    // Broker
    const char* host = nullptr;
    uint16_t    port = 1883;

    // Auth
    const char* clientId = nullptr;
    const char* user = nullptr;
    const char* password = nullptr;

    // Last Will and Testament (optional)
    // If lwtTopic is null/empty, LWT should be considered disabled by the client implementation.
    const char* lwtTopic = nullptr;
    const char* lwtPayloadOnline = nullptr;
    const char* lwtPayloadOffline = nullptr;
};

class MqttClient {
public:
    static bool init(const MqttConfig& cfg);
    static void loop();

    static bool publish(const char* topic, const char* payload, bool retained = false);
    static bool isConnected();
};