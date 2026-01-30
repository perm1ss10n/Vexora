#include "cmd/cmd_processor.h"

#include <Arduino.h>
#include <cstdio>

#include <ArduinoJson.h>

#include "log/logger.h"
#include "mqtt/mqtt_client.h"
#include "mqtt/mqtt_topics.h"
#include "state/state_publisher.h"

const char* CommandProcessor::g_deviceId = nullptr;

void CommandProcessor::init(const char* deviceId) {
    g_deviceId = deviceId;
    LOGI("CMD", "init()");
}

static bool topicIsCmdForThisDevice(const char* topic, const char* deviceId) {
    if (!topic || !deviceId || deviceId[0] == '\0') return false;

    // expected: v1/dev/<deviceId>/cmd
    char expected[96];
    if (!vx_build_topic(expected, sizeof(expected), deviceId, VX_T_CMD)) return false;

    return strcmp(topic, expected) == 0;
}

void CommandProcessor::onMessage(const char* topic, const uint8_t* payload, size_t len) {
    if (!g_deviceId || g_deviceId[0] == '\0') return;
    if (!topicIsCmdForThisDevice(topic, g_deviceId)) return;

    if (!payload || len == 0) {
        sendAck("", false, "BAD_JSON", "empty payload");
        return;
    }

    if (!handleCmd(payload, len)) {
        // handleCmd already acks on parse errors
        return;
    }
}

bool CommandProcessor::handleCmd(const uint8_t* payload, size_t len) {
    StaticJsonDocument<512> doc;

    // ArduinoJson expects a mutable char* sometimes; but deserializeJson accepts uint8_t*
    DeserializationError err = deserializeJson(doc, payload, len);
    if (err) {
        sendAck("", false, "BAD_JSON", err.c_str());
        return false;
    }

    const int v = doc["v"] | 0;
    const char* id = doc["id"] | "";
    const char* type = doc["type"] | "";

    if (v != 1) {
        sendAck(id, false, "BAD_VERSION", "unsupported v");
        return false;
    }
    if (!type || type[0] == '\0') {
        sendAck(id, false, "BAD_CMD", "missing type");
        return false;
    }

    // MVP commands
    if (strcmp(type, "ping") == 0) {
        sendAck(id, true, "OK", "");
        sendEvent("PING", "pong");
        return true;
    }

    if (strcmp(type, "get_state") == 0) {
        // просто форсим publish retained state
        StatePublisher::markDirty();
        sendAck(id, true, "OK", "");
        sendEvent("GET_STATE", "published");
        return true;
    }

    if (strcmp(type, "reboot") == 0) {
        sendAck(id, true, "OK", "rebooting");
        sendEvent("REBOOT", "requested");

        delay(150); // дать MQTT отдать пакет
        ESP.restart();
        return true;
    }

    sendAck(id, false, "UNKNOWN_CMD", type);
    return false;
}

void CommandProcessor::sendAck(const char* id, bool ok, const char* code, const char* msg) {
    if (!g_deviceId || g_deviceId[0] == '\0') return;

    char topic[96];
    if (!vx_build_topic(topic, sizeof(topic), g_deviceId, VX_T_ACK)) return;

    const uint64_t ts = (uint64_t)time(nullptr) * 1000ULL; // если время не установлено — будет 0, для MVP ок
    char payload[384];

    snprintf(payload, sizeof(payload),
             "{\"v\":1,\"id\":\"%s\",\"deviceId\":\"%s\",\"ts\":%llu,\"ok\":%s,\"code\":\"%s\",\"msg\":\"%s\"}",
             id ? id : "",
             g_deviceId,
             (unsigned long long)ts,
             ok ? "true" : "false",
             code ? code : "",
             msg ? msg : "");

    (void)MqttClient::publish(topic, payload, false);
}

void CommandProcessor::sendEvent(const char* code, const char* msg) {
    if (!g_deviceId || g_deviceId[0] == '\0') return;

    char topic[96];
    if (!vx_build_topic(topic, sizeof(topic), g_deviceId, VX_T_EVENT)) return;

    const uint64_t ts = (uint64_t)time(nullptr) * 1000ULL;
    char payload[256];

    snprintf(payload, sizeof(payload),
             "{\"v\":1,\"deviceId\":\"%s\",\"ts\":%llu,\"code\":\"%s\",\"msg\":\"%s\"}",
             g_deviceId,
             (unsigned long long)ts,
             code ? code : "",
             msg ? msg : "");

    (void)MqttClient::publish(topic, payload, false);
}