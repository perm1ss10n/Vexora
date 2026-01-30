#pragma once

#include <stddef.h>
#include <stdint.h>

class CommandProcessor {
public:
    static void init(const char* deviceId);

    // raw payload from MQTT callback
    static void onMessage(const char* topic, const uint8_t* payload, size_t len);

private:
    static bool handleCmd(const uint8_t* payload, size_t len);

    static void sendAck(const char* id, bool ok, const char* code, const char* msg);
    static void sendEvent(const char* code, const char* msg);

    static const char* g_deviceId;
};