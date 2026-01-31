#include "config/config_manager.h"

#include <Arduino.h>
#include <cstdio>

#include "log/logger.h"
#include "mqtt/mqtt_client.h"
#include "mqtt/mqtt_topics.h"
#include "telemetry/telemetry.h"

// ===== static storage =====
DeviceConfig ConfigManager::g_active{};
DeviceConfig ConfigManager::g_previous{};
DeviceConfig ConfigManager::g_pending{};
bool ConfigManager::g_hasPending = false;

const char *ConfigManager::g_deviceId = nullptr;

uint32_t ConfigManager::g_pendingDeadlineMs = 0;
bool ConfigManager::g_pendingSawMqtt = false;

static inline bool nonEmpty(const char *s)
{
    return s && s[0] != '\0';
}

static inline void setRes(ApplyCfgResult &r, bool ok, const char *code, const char *msg)
{
    r.ok = ok;
    r.code = code;
    r.msg = msg;
}

bool ConfigManager::init()
{
    g_active.valid = true;

    // ===== DEFAULTS (MVP) =====
    // Broker пока “временный”.
    g_active.mqtt.host = "broker.hivemq.com"; // временно
    g_active.mqtt.port = 1883;
    g_active.mqtt.clientId = "vexora-dev";
    g_active.mqtt.user = "";
    g_active.mqtt.password = "";

    // mqtt_client сам умеет собрать v1/dev/<deviceId>/lwt если lwtTopic == nullptr
    g_active.mqtt.lwtTopic = nullptr;
    g_active.mqtt.lwtPayloadOnline = "online";
    g_active.mqtt.lwtPayloadOffline = "offline";

    g_active.telemetry.intervalMs = 5000;
    g_active.telemetry.minPublishMs = 0;

    g_previous = g_active;

    LOGI("CFG", "config loaded (defaults)");
    return true;
}

void ConfigManager::setDeviceId(const char *deviceId)
{
    g_deviceId = deviceId;

    // Пробрасываем deviceId в active mqtt-конфиг
    g_active.mqtt.deviceId = g_deviceId;
    g_active.mqtt.clientId = g_deviceId;

    // На случай если был pending — тоже обновим
    if (g_hasPending)
    {
        g_pending.mqtt.deviceId = g_deviceId;
        g_pending.mqtt.clientId = g_deviceId;
    }
}

DeviceConfig ConfigManager::getActive()
{
    return g_active; // копия
}

ApplyCfgResult ConfigManager::applyCandidate(JsonObject cfg)
{
    ApplyCfgResult res{};
    setRes(res, false, "CFG_REJECTED", "validation failed");

    DeviceConfig candidate{};

    if (!validateCandidate(g_active, cfg, candidate, res))
    {
        publishCfgStatus("rejected");
        return res;
    }

    // pending apply
    g_pending = candidate;
    g_hasPending = true;
    g_pendingSawMqtt = false;

    // применяем runtime сразу (двухфазно)
    applyRuntime(g_pending);
    startPendingWindow();

    publishCfgStatus("pending");

    ApplyCfgResult okRes{};
    setRes(okRes, true, "OK", "cfg pending applied");
    return okRes;
}

bool ConfigManager::validateCandidate(const DeviceConfig &base, JsonObject cfg, DeviceConfig &outCandidate, ApplyCfgResult &outRes)
{
    // partial cfg allowed
    outCandidate = base;

    // ---- telemetry ----
    if (cfg.containsKey("telemetry"))
    {
        JsonVariant telemetryV = cfg["telemetry"]; // JsonVariant supports is<T>()
        if (!telemetryV.is<JsonObject>())
        {
            setRes(outRes, false, "BAD_CFG", "telemetry must be object");
            return false;
        }

        JsonObject t = telemetryV.as<JsonObject>();

        // intervalMs обязателен, если telemetry блок присутствует
        if (!t.containsKey("intervalMs"))
        {
            setRes(outRes, false, "BAD_CFG", "telemetry.intervalMs is required");
            return false;
        }
        if (!t["intervalMs"].is<uint32_t>())
        {
            setRes(outRes, false, "BAD_CFG", "telemetry.intervalMs must be uint32");
            return false;
        }

        uint32_t interval = t["intervalMs"].as<uint32_t>();
        if (interval < 1000 || interval > 3600000)
        {
            setRes(outRes, false, "CFG_REJECTED", "intervalMs out of range (1000..3600000)");
            return false;
        }
        outCandidate.telemetry.intervalMs = interval;

        // minPublishMs опционально
        if (t.containsKey("minPublishMs"))
        {
            if (!t["minPublishMs"].is<uint32_t>())
            {
                setRes(outRes, false, "BAD_CFG", "telemetry.minPublishMs must be uint32");
                return false;
            }
            outCandidate.telemetry.minPublishMs = t["minPublishMs"].as<uint32_t>();
        }
    }

    // ---- mqtt ----
    // MVP: пока запрещаем менять broker/port/cred по сети, чтобы не словить кирпич.

    setRes(outRes, true, "OK", "validated");
    return true;
}

void ConfigManager::applyRuntime(const DeviceConfig &cfg)
{
    // Применяем то, что можно менять на лету в MVP
    Telemetry::updateInterval(cfg.telemetry.intervalMs);
    Telemetry::updateMinPublishMs(cfg.telemetry.minPublishMs);
}

void ConfigManager::publishCfgStatus(const char *status)
{
    if (!nonEmpty(g_deviceId))
        return;
    if (!MqttClient::isConnected())
        return;

    char topic[96];
    if (!vx_build_topic(topic, sizeof(topic), g_deviceId, VX_T_CFG_STATUS))
        return;

    char payload[160];
    snprintf(payload, sizeof(payload),
             "{\"v\":1,\"deviceId\":\"%s\",\"status\":\"%s\"}",
             g_deviceId,
             status ? status : "");

    (void)MqttClient::publish(topic, payload, true);
}

void ConfigManager::startPendingWindow()
{
    const uint32_t now = millis();
    g_pendingDeadlineMs = now + PENDING_GRACE_MS;

    char buf[96];
    snprintf(buf, sizeof(buf), "pending window %ums", (unsigned)PENDING_GRACE_MS);
    LOGI("CFG", buf);
}

void ConfigManager::commitPending()
{
    g_previous = g_active;
    g_active = g_pending;
    g_hasPending = false;

    publishCfgStatus("applied");
    LOGI("CFG", "commit pending");
}

void ConfigManager::rollbackPending()
{
    g_active = g_previous;
    g_hasPending = false;

    applyRuntime(g_active);

    publishCfgStatus("rolled_back");
    LOGW("CFG", "rollback pending");
}

void ConfigManager::loop()
{
    if (!g_hasPending)
        return;

    // health: MQTT должен стать connected хотя бы раз в grace window
    if (MqttClient::isConnected())
    {
        g_pendingSawMqtt = true;
    }

    const uint32_t now = millis();
    if (g_pendingDeadlineMs != 0 && now < g_pendingDeadlineMs)
    {
        return;
    }

    // дедлайн истёк
    if (g_pendingSawMqtt)
    {
        commitPending();
    }
    else
    {
        rollbackPending();
    }
}