#include "offline/offline_queue.h"

#include <cstring>
#include <cstdio>

#include "log/logger.h"
#include "mqtt/mqtt_client.h"

static OfflineMessage* g_buf = nullptr;
static size_t g_capacity = 0;
static size_t g_head = 0;
static size_t g_tail = 0;
static size_t g_count = 0;

void OfflineQueue::init(size_t capacity)
{
    static OfflineMessage storage[32]; // жёсткий лимит MVP
    g_buf = storage;
    g_capacity = capacity > 32 ? 32 : capacity;
    g_head = g_tail = g_count = 0;

    char buf[64];
    snprintf(buf, sizeof(buf), "init cap=%u", (unsigned)g_capacity);
    LOGI("OFFQ", buf);
}

bool OfflineQueue::push(const char* topic,
                        const char* payload,
                        bool retained)
{
    if (!topic || !payload || !g_buf)
        return false;

    // если очередь полная — дропаем самое старое
    if (g_count == g_capacity) {
        g_tail = (g_tail + 1) % g_capacity;
        g_count--;
        LOGW("OFFQ", "overflow, dropping oldest");
    }

    OfflineMessage& m = g_buf[g_head];

    strncpy(m.topic, topic, sizeof(m.topic) - 1);
    m.topic[sizeof(m.topic) - 1] = '\0';

    strncpy(m.payload, payload, sizeof(m.payload) - 1);
    m.payload[sizeof(m.payload) - 1] = '\0';

    m.retained = retained;

    g_head = (g_head + 1) % g_capacity;
    g_count++;

    return true;
}

void OfflineQueue::flush()
{
    if (!MqttClient::isConnected() || g_count == 0)
        return;

    while (g_count > 0) {
        OfflineMessage& m = g_buf[g_tail];
        if (!MqttClient::publish(m.topic, m.payload, m.retained)) {
            // если publish не удался — выходим, попробуем позже
            LOGW("OFFQ", "flush failed, retry later");
            return;
        }
        g_tail = (g_tail + 1) % g_capacity;
        g_count--;
    }

    LOGI("OFFQ", "flushed");
}

size_t OfflineQueue::size()
{
    return g_count;
}