#include "mqtt/mqtt_client.h"
#include "log/logger.h"

#include <Arduino.h>
#include <PubSubClient.h>
#include <cstdio>

static char g_lwtTopicBuf[96];

// PubSubClient singleton
static PubSubClient* g_client = nullptr;

// ВАЖНО: храним копию конфига внутри модуля (чтобы setDeviceId был безопасным)
static MqttConfig g_cfgCopy;
static const MqttConfig* g_cfg = nullptr;

static MqttClient::MessageCallback g_msgCb = nullptr;

// Reconnect backoff
static uint32_t g_nextRetryMs = 0;
static uint32_t g_backoffMs = 1000;

static const uint32_t BACKOFF_MIN_MS = 1000;
static const uint32_t BACKOFF_MAX_MS = 30000;

static uint32_t g_reconnectCount = 0;

static inline bool isNonEmpty(const char* s) {
    return s && s[0] != '\0';
}

static const char* resolveLwtTopic(const MqttConfig& cfg) {
    // 1) Explicit topic wins
    if (isNonEmpty(cfg.lwtTopic)) return cfg.lwtTopic;

    // 2) If deviceId is known, build: v1/dev/<deviceId>/lwt
    if (!isNonEmpty(cfg.deviceId)) return nullptr;

    if (vx_build_topic(g_lwtTopicBuf, sizeof(g_lwtTopicBuf), cfg.deviceId, VX_T_LWT)) {
        return g_lwtTopicBuf;
    }

    return nullptr;
}

static void onMessage(char* topic, uint8_t* payload, unsigned int len) {
    const char* t = (topic && topic[0] != '\0') ? topic : nullptr;

    if (g_msgCb) {
        g_msgCb(t, payload, (size_t)len);
        return;
    }

    char buf[96];
    snprintf(buf, sizeof(buf), "recv topic=%s len=%u", t ? t : "(null)", (unsigned)len);
    LOGI("MQTT", buf);
}

static void subscribeTopics() {
    if (!g_client) return;
    if (!g_cfg || !isNonEmpty(g_cfg->deviceId)) {
        LOGW("MQTT", "subscribe skipped: deviceId not set");
        return;
    }

    char topic[96];

    if (vx_build_topic(topic, sizeof(topic), g_cfg->deviceId, VX_T_CMD)) g_client->subscribe(topic, 1);
    if (vx_build_topic(topic, sizeof(topic), g_cfg->deviceId, VX_T_CFG)) g_client->subscribe(topic, 1);
    if (vx_build_topic(topic, sizeof(topic), g_cfg->deviceId, VX_T_OTA)) g_client->subscribe(topic, 1);
}

static bool tryConnect() {
    if (!g_client || !g_cfg) return false;

    const MqttConfig& cfg = *g_cfg;

    if (!isNonEmpty(cfg.host)) {
        LOGE("MQTT", "cfg.host is empty");
        return false;
    }

    g_client->setServer(cfg.host, cfg.port);

    const char* clientId = isNonEmpty(cfg.clientId) ? cfg.clientId : "vexora";
    const char* user     = isNonEmpty(cfg.user) ? cfg.user : nullptr;
    const char* pass     = isNonEmpty(cfg.password) ? cfg.password : nullptr;

    const char* lwtTopic = resolveLwtTopic(cfg);

    bool ok = false;

    // Enable LWT if we have a topic and an offline payload
    if (lwtTopic && isNonEmpty(cfg.lwtPayloadOffline)) {
        // willQoS=1, willRetain=true
        ok = g_client->connect(clientId, user, pass, lwtTopic, 1, true, cfg.lwtPayloadOffline);
    } else {
        ok = g_client->connect(clientId, user, pass);
    }

    if (!ok) {
        int st = g_client->state();
        char buf[64];
        snprintf(buf, sizeof(buf), "connect failed state=%d", st);
        LOGW("MQTT", buf);
        return false;
    }

    // Connected -> reset backoff
    g_backoffMs = BACKOFF_MIN_MS;
    g_nextRetryMs = 0;

    // Подписки (3.2)
    subscribeTopics();

    // If online payload is configured — publish retained to the same LWT topic
    if (lwtTopic && isNonEmpty(cfg.lwtPayloadOnline)) {
        g_client->publish(lwtTopic, cfg.lwtPayloadOnline, true);
    }

    LOGI("MQTT", "connected");
    return true;
}

bool MqttClient::init(const MqttConfig& cfg, Client& netClient) {
    static PubSubClient client(netClient);
    g_client = &client;

    // копируем конфиг внутрь
    g_cfgCopy = cfg;
    g_cfg = &g_cfgCopy;

    g_nextRetryMs = 0;
    g_backoffMs = BACKOFF_MIN_MS;
    g_reconnectCount = 0;

    g_client->setKeepAlive(30);
    g_client->setCallback(onMessage);

    LOGI("MQTT", "init()");
    return true;
}

void MqttClient::loop() {
    if (!g_client || !g_cfg) return;

    if (g_client->connected()) {
        g_client->loop();
        return;
    }

    const uint32_t now = millis();
    if (g_nextRetryMs != 0 && now < g_nextRetryMs) return;

    if (tryConnect()) return;

    // schedule next retry with exponential backoff
    g_reconnectCount++;

    g_nextRetryMs = now + g_backoffMs;

    if (g_backoffMs < BACKOFF_MAX_MS) {
        g_backoffMs *= 2;
        if (g_backoffMs > BACKOFF_MAX_MS) g_backoffMs = BACKOFF_MAX_MS;
    }

    char buf[96];
    snprintf(buf, sizeof(buf), "reconnect in %ums (count=%lu)",
             (unsigned)g_backoffMs, (unsigned long)g_reconnectCount);
    LOGW("MQTT", buf);
}

bool MqttClient::publish(const char* topic, const char* payload, bool retained) {
    if (!g_client || !g_client->connected()) return false;
    if (!isNonEmpty(topic)) return false;

    if (!payload) payload = "";
    return g_client->publish(topic, payload, retained);
}

bool MqttClient::isConnected() {
    return g_client && g_client->connected();
}

uint32_t MqttClient::reconnectCount() {
    return g_reconnectCount;
}

void MqttClient::setMessageCallback(MessageCallback cb) {
    g_msgCb = cb;
}

void MqttClient::setDeviceId(const char* deviceId) {
    g_cfgCopy.deviceId = deviceId;
    g_cfg = &g_cfgCopy;
}