#pragma once

#include <stdint.h>

struct StatePublishConfig
{
    uint32_t intervalMs = 5000; // периодический publish (на всякий)
};

class StatePublisher
{
public:
    static void init(const StatePublishConfig &cfg, const char *deviceId, const char *fwVersion);
    static void loop();

    // можно дергать вручную при смене состояния
    static void setStatus(const char *status); // "online/offline/degraded/error"
    static void setLink(const char *linkType); // "wifi/gsm"
    static void setIP(const char *ip);         // для wifi
    static void setRssi(int rssi);             // -127..0 (wifi), 0 если неизвестно
    static void markDirty();                   // форс-публикация

private:
    static void publishNow(bool force);
};