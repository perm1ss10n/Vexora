#include "telemetry/telemetry.h"

#include <Arduino.h>
#include <cstdio>
#include <ctime>

#include "log/logger.h"
#include "mqtt/mqtt_client.h"
#include "mqtt/mqtt_topics.h"
#include "offline/offline_queue.h"

static TelemetryConfig g_cfg;
static const char *g_deviceId = nullptr;

static uint32_t g_nextPublishMs = 0;
static uint32_t g_lastPublishMs = 0;

static char g_topic[96];

static bool buildTopicOnce(const char *deviceId)
{
    if (!deviceId || deviceId[0] == '\0')
        return false;
    g_topic[0] = '\0';
    return vx_build_topic(g_topic, sizeof(g_topic), deviceId, VX_T_TELEMETRY);
}

// Epoch ms if time is set; otherwise fallback to millis() (monotonic, ok for MVP)
static uint64_t nowMs()
{
    time_t t = time(nullptr);
    // If SNTP/RTC not set, many ESP32 setups return 0 or a small value.
    if (t < 1672531200) // 2023-01-01
        return (uint64_t)millis();
    return (uint64_t)t * 1000ULL;
}

void Telemetry::init(const TelemetryConfig &cfg, const char *deviceId)
{
    g_cfg = cfg;
    g_deviceId = deviceId;

    // Defaults (safety)
    if (g_cfg.intervalMs == 0)
        g_cfg.intervalMs = 5000;

    g_nextPublishMs = 0;
    g_lastPublishMs = 0;

    if (!buildTopicOnce(g_deviceId))
    {
        LOGW("TEL", "topic build failed (deviceId missing?)");
    }

    LOGI("TEL", "init()");
}

void Telemetry::loop()
{
    if (!g_deviceId || g_deviceId[0] == '\0')
        return;

    const uint32_t now = millis();
    if (g_nextPublishMs != 0 && now < g_nextPublishMs)
        return;

    // В MVP: отправляем только когда MQTT подключён (иначе просто молчим)
    if (!MqttClient::isConnected())
    {
        // Перекидываем следующий тик чуть позже
        g_nextPublishMs = now + g_cfg.intervalMs;
        return;
    }
    // ensure topic is built (in case deviceId appeared later)
    if (g_topic[0] == '\0')
    {
        (void)buildTopicOnce(g_deviceId);
        if (g_topic[0] == '\0')
        { // всё ещё не собрался
            g_nextPublishMs = now + g_cfg.intervalMs;
            return;
        }
    }

    publishTick();
    g_nextPublishMs = now + g_cfg.intervalMs;
}

bool Telemetry::publishMetric(const char *key, float value, const char *unit)
{
    if (!g_deviceId || g_deviceId[0] == '\0')
        return false;
    if (!MqttClient::isConnected())
        return false;
    if (!key || key[0] == '\0')
        return false;

    // topic (cached)
    if (g_topic[0] == '\0')
        return false;

    // payload (MVP: 1 metric)
    // {"v":1,"deviceId":"...","ts":123,"metrics":[{"key":"k","value":1.23,"unit":"c"}]}
    char payload[256];
    const uint64_t ts = nowMs();
    if (unit && unit[0] != '\0')
    {
        snprintf(payload, sizeof(payload),
                 "{\"v\":1,\"deviceId\":\"%s\",\"ts\":%llu,\"metrics\":[{\"key\":\"%s\",\"value\":%.4f,\"unit\":\"%s\"}]}",
                 g_deviceId,
                 (unsigned long long)ts,
                 key,
                 (double)value,
                 unit);
    }
    else
    {
        snprintf(payload, sizeof(payload),
                 "{\"v\":1,\"deviceId\":\"%s\",\"ts\":%llu,\"metrics\":[{\"key\":\"%s\",\"value\":%.4f}]}",
                 g_deviceId,
                 (unsigned long long)ts,
                 key,
                 (double)value);
    }

    if (MqttClient::isConnected())
    {
        return MqttClient::publish(g_topic, payload, false);
    }

    // offline → в очередь
    OfflineQueue::push(g_topic, payload, false);
    return true;
}

void Telemetry::publishTick()
{
    const uint32_t now = millis();

    // защита от слишком частых публикаций
    if (g_cfg.minPublishMs > 0 && g_lastPublishMs != 0 && (now - g_lastPublishMs) < g_cfg.minPublishMs)
    {
        return;
    }
    g_lastPublishMs = now;

    // topic (cached)
    if (g_topic[0] == '\0')
        return;

    // MVP: шлём uptime_sec как одну метрику
    const uint64_t ts = nowMs();

    char payload[256];
    snprintf(payload, sizeof(payload),
             "{\"v\":1,\"deviceId\":\"%s\",\"ts\":%llu,\"metrics\":[{\"key\":\"uptime_sec\",\"value\":%.0f,\"unit\":\"s\"}]}",
             g_deviceId,
             (unsigned long long)ts,
             (double)(millis() / 1000.0));

    if (MqttClient::isConnected())
    {
        (void)MqttClient::publish(g_topic, payload, false);
    }
    else
    {
        OfflineQueue::push(g_topic, payload, false);
    }
}
void Telemetry::updateInterval(uint32_t intervalMs)
{
    if (intervalMs == 0)
        return;

    g_cfg.intervalMs = intervalMs;

    // пересчитываем следующий тик
    g_nextPublishMs = millis() + g_cfg.intervalMs;

    LOGI("TEL", "interval updated");
}