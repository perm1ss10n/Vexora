#include "state/state_publisher.h"

#include <Arduino.h>

#include "log/logger.h"
#include "mqtt/mqtt_client.h"
#include "mqtt/mqtt_topics.h"

static StatePublishConfig g_cfg;
static const char *g_deviceId = nullptr;
static const char *g_fw = nullptr;

static char g_status[16] = "offline";
static char g_link[8] = "";
static char g_ip[32] = "";
static int g_rssi = 0;

static uint32_t g_nextPublishMs = 0;
static bool g_dirty = true;

static bool buildTopic(char *out, size_t outSize, const char *deviceId)
{
    return vx_build_topic(out, outSize, deviceId, VX_T_STATE);
}

void StatePublisher::init(const StatePublishConfig &cfg, const char *deviceId, const char *fwVersion)
{
    g_cfg = cfg;
    g_deviceId = deviceId;
    g_fw = fwVersion;

    g_nextPublishMs = 0;
    g_dirty = true;

    LOGI("STATE", "init()");
}

void StatePublisher::setStatus(const char *status)
{
    if (!status || status[0] == '\0')
        return;
    strncpy(g_status, status, sizeof(g_status) - 1);
    g_status[sizeof(g_status) - 1] = '\0';
    g_dirty = true;
}

void StatePublisher::setLink(const char *linkType)
{
    if (!linkType)
        linkType = "";
    strncpy(g_link, linkType, sizeof(g_link) - 1);
    g_link[sizeof(g_link) - 1] = '\0';
    g_dirty = true;
}

void StatePublisher::setIP(const char *ip)
{
    if (!ip)
        ip = "";
    strncpy(g_ip, ip, sizeof(g_ip) - 1);
    g_ip[sizeof(g_ip) - 1] = '\0';
    g_dirty = true;
}

void StatePublisher::setRssi(int rssi)
{
    g_rssi = rssi;
    g_dirty = true;
}

void StatePublisher::markDirty()
{
    g_dirty = true;
}

void StatePublisher::loop()
{
    if (!g_deviceId || g_deviceId[0] == '\0')
        return;

    const uint32_t now = millis();
    if (g_nextPublishMs != 0 && now < g_nextPublishMs)
        return;

    // состояние публикуем только при наличии MQTT
    if (!MqttClient::isConnected())
    {
        // но статус можно обновить локально
        setStatus("offline");
        g_nextPublishMs = now + g_cfg.intervalMs;
        return;
    }

    // когда MQTT поднят — online
    setStatus("online");

    publishNow(false);
    g_nextPublishMs = now + g_cfg.intervalMs;
}

void StatePublisher::publishNow(bool force)
{
    if (!MqttClient::isConnected())
        return;
    if (!force && !g_dirty)
        return;

    char topic[96];
    if (!buildTopic(topic, sizeof(topic), g_deviceId))
        return;

    // retained=true (важно)
    // {"v":1,"deviceId":"...","ts":...,"status":"online","link":{"type":"wifi","rssi":-55,"ip":"..."},"fw":"...","uptimeSec":123}
    char payload[384];

    const uint64_t ts = (uint64_t)time(nullptr) * 1000ULL;
    const uint32_t uptimeSec = millis() / 1000U;

    // link object (минимально)
    // если ip пустой — не вставляем ip
    if (g_ip[0] != '\0')
    {
        snprintf(payload, sizeof(payload),
                 "{\"v\":1,\"deviceId\":\"%s\",\"ts\":%llu,\"status\":\"%s\",\"link\":{\"type\":\"%s\",\"rssi\":%d,\"ip\":\"%s\"},\"fw\":\"%s\",\"uptimeSec\":%u}",
                 g_deviceId,
                 (unsigned long long)ts,
                 g_status,
                 (g_link[0] ? g_link : "wifi"),
                 g_rssi,
                 g_ip,
                 (g_fw ? g_fw : ""),
                 (unsigned)uptimeSec);
    }
    else
    {
        snprintf(payload, sizeof(payload),
                 "{\"v\":1,\"deviceId\":\"%s\",\"ts\":%llu,\"status\":\"%s\",\"link\":{\"type\":\"%s\",\"rssi\":%d},\"fw\":\"%s\",\"uptimeSec\":%u}",
                 g_deviceId,
                 (unsigned long long)ts,
                 g_status,
                 (g_link[0] ? g_link : "wifi"),
                 g_rssi,
                 (g_fw ? g_fw : ""),
                 (unsigned)uptimeSec);
    }

    (void)MqttClient::publish(topic, payload, true);
    g_dirty = false;
}