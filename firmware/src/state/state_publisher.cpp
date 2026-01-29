#include "state/state_publisher.h"

#include <Arduino.h>

#include "log/logger.h"
#include "mqtt/mqtt_client.h"
#include "mqtt/mqtt_topics.h"

#include <ctime>
#include <cstdio>
#include <cstring>

static StatePublishConfig g_cfg;
static const char *g_deviceId = nullptr;
static const char *g_fw = nullptr;

static char g_status[16] = "offline";
static char g_link[8] = "";
static char g_ip[32] = "";
static int g_rssi = 0;

static uint32_t g_nextPublishMs = 0;
static bool g_dirty = true;

static uint64_t nowMs()
{
    time_t t = time(nullptr);
    if (t < 1672531200)
        return (uint64_t)millis();
    return (uint64_t)t * 1000ULL;
}

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
    if (strncmp(g_status, status, sizeof(g_status)) == 0)
        return;

    strncpy(g_status, status, sizeof(g_status) - 1);
    g_status[sizeof(g_status) - 1] = '\0';
    g_dirty = true;
}

void StatePublisher::setLink(const char *linkType)
{
    if (!linkType)
        linkType = "";
    if (strncmp(g_link, linkType, sizeof(g_link)) == 0)
        return;

    strncpy(g_link, linkType, sizeof(g_link) - 1);
    g_link[sizeof(g_link) - 1] = '\0';
    g_dirty = true;
}

void StatePublisher::setIP(const char *ip)
{
    if (!ip)
        ip = "";
    if (strncmp(g_ip, ip, sizeof(g_ip)) == 0)
        return;

    strncpy(g_ip, ip, sizeof(g_ip) - 1);
    g_ip[sizeof(g_ip) - 1] = '\0';
    g_dirty = true;
}

void StatePublisher::setRssi(int rssi)
{
    if (g_rssi == rssi)
        return;
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

    if (!MqttClient::isConnected())
    {
        g_nextPublishMs = now + g_cfg.intervalMs;
        return;
    }

    // Периодический publish: даже если ничего не менялось (retained refresh)
    publishNow(true);
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

    const uint64_t ts = nowMs();
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
                 (g_link[0] ? g_link : "unknown"),
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
                 (g_link[0] ? g_link : "unknown"),
                 g_rssi,
                 (g_fw ? g_fw : ""),
                 (unsigned)uptimeSec);
    }

    (void)MqttClient::publish(topic, payload, true);
    g_dirty = false;
}