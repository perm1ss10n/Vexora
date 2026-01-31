#include "cmd/cmd_processor.h"

#include <Arduino.h>
#include <cstdio>
#include <ctime>

#include <ArduinoJson.h>

#include "log/logger.h"
#include "config/config_manager.h"
#include "mqtt/mqtt_client.h"
#include "mqtt/mqtt_topics.h"
#include "state/state_publisher.h"

const char *CommandProcessor::g_deviceId = nullptr;

// ===== init =====

void CommandProcessor::init(const char *deviceId)
{
    g_deviceId = deviceId;
    LOGI("CMD", "init()");
}

// ===== helpers =====

static bool topicIsCmdForThisDevice(const char *topic, const char *deviceId)
{
    if (!topic || !deviceId || deviceId[0] == '\0')
        return false;

    char expected[96];
    if (!vx_build_topic(expected, sizeof(expected), deviceId, VX_T_CMD))
        return false;

    return strcmp(topic, expected) == 0;
}

static void publishCfgStatus(const char *deviceId, const char *status)
{
    if (!deviceId || deviceId[0] == '\0')
        return;

    char topic[96];
    if (!vx_build_topic(topic, sizeof(topic), deviceId, VX_T_CFG_STATUS))
        return;

    char payload[128];
    snprintf(payload, sizeof(payload),
             "{\"v\":1,\"deviceId\":\"%s\",\"status\":\"%s\"}",
             deviceId,
             status ? status : "");

    (void)MqttClient::publish(topic, payload, true);
}

// ===== MQTT entry =====

void CommandProcessor::onMessage(const char *topic, const uint8_t *payload, size_t len)
{
    if (!g_deviceId || !topicIsCmdForThisDevice(topic, g_deviceId))
        return;

    if (!payload || len == 0)
    {
        sendAck("", false, "BAD_JSON", "empty payload");
        return;
    }

    handleCmd(payload, len);
}

// ===== command handling =====

bool CommandProcessor::handleCmd(const uint8_t *payload, size_t len)
{
    StaticJsonDocument<512> doc;
    DeserializationError err = deserializeJson(doc, payload, len);
    if (err)
    {
        sendAck("", false, "BAD_JSON", err.c_str());
        return false;
    }

    const int v = doc["v"] | 0;
    const char *id = doc["id"] | "";
    const char *type = doc["type"] | "";

    if (v != 1)
    {
        sendAck(id, false, "BAD_VERSION", "unsupported v");
        return false;
    }

    if (!type || type[0] == '\0')
    {
        sendAck(id, false, "BAD_CMD", "missing type");
        return false;
    }

    // ===== MVP commands =====

    if (strcmp(type, "ping") == 0)
    {
        sendAck(id, true, "OK", "");
        sendEvent("PING", "pong");
        return true;
    }

    if (strcmp(type, "get_state") == 0)
    {
        StatePublisher::markDirty();
        sendAck(id, true, "OK", "");
        sendEvent("GET_STATE", "published");
        return true;
    }

    if (strcmp(type, "reboot") == 0)
    {
        sendAck(id, true, "OK", "rebooting");
        sendEvent("REBOOT", "requested");
        delay(150);
        ESP.restart();
        return true;
    }

    if (strcmp(type, "apply_cfg") == 0)
    {
        JsonObject cfg = doc["cfg"];
        if (cfg.isNull())
        {
            sendAck(id, false, "BAD_CFG", "cfg missing");
            publishCfgStatus(g_deviceId, "rejected");
            return false;
        }

        ApplyCfgResult r = ConfigManager::applyCandidate(cfg);
        if (!r.ok)
        {
            sendAck(id, false,
                    r.code ? r.code : "CFG_REJECTED",
                    r.msg ? r.msg : "validation failed");
            publishCfgStatus(g_deviceId, "rejected");
            return false;
        }

        sendAck(id, true, "OK", r.msg ? r.msg : "cfg applied");
        sendEvent("CFG_APPLIED", "");
        publishCfgStatus(g_deviceId, "applied");
        return true;
    }

    sendAck(id, false, "UNKNOWN_CMD", type);
    return false;
}

// ===== ACK / EVENT =====

void CommandProcessor::sendAck(const char *id, bool ok, const char *code, const char *msg)
{
    if (!g_deviceId)
        return;

    char topic[96];
    if (!vx_build_topic(topic, sizeof(topic), g_deviceId, VX_T_ACK))
        return;

    const uint64_t ts = (uint64_t)time(nullptr) * 1000ULL;
    char payload[384];

    snprintf(payload, sizeof(payload),
             "{\"v\":1,\"id\":\"%s\",\"deviceId\":\"%s\",\"ts\":%llu,"
             "\"ok\":%s,\"code\":\"%s\",\"msg\":\"%s\"}",
             id ? id : "",
             g_deviceId,
             (unsigned long long)ts,
             ok ? "true" : "false",
             code ? code : "",
             msg ? msg : "");

    (void)MqttClient::publish(topic, payload, false);
}

void CommandProcessor::sendEvent(const char *code, const char *msg)
{
    if (!g_deviceId)
        return;

    char topic[96];
    if (!vx_build_topic(topic, sizeof(topic), g_deviceId, VX_T_EVENT))
        return;

    const uint64_t ts = (uint64_t)time(nullptr) * 1000ULL;
    char payload[256];

    snprintf(payload, sizeof(payload),
             "{\"v\":1,\"deviceId\":\"%s\",\"ts\":%llu,"
             "\"code\":\"%s\",\"msg\":\"%s\"}",
             g_deviceId,
             (unsigned long long)ts,
             code ? code : "",
             msg ? msg : "");

    (void)MqttClient::publish(topic, payload, false);
}