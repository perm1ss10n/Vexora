#pragma once
#include <stdint.h>
#include <stddef.h>

struct OfflineMessage {
    char topic[96];
    char payload[256];
    bool retained;
};

class OfflineQueue {
public:
    static void init(size_t capacity = 20);

    static bool push(const char* topic,
                     const char* payload,
                     bool retained);

    static void flush();
    static size_t size();
};