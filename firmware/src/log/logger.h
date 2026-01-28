#pragma once
#include <Arduino.h>

class Logger {
public:
    static void init();
    static void info(const char* tag, const char* msg);
    static void warn(const char* tag, const char* msg);
    static void error(const char* tag, const char* msg);
};

#define LOGI(tag, msg) Logger::info(tag, msg)
#define LOGW(tag, msg) Logger::warn(tag, msg)
#define LOGE(tag, msg) Logger::error(tag, msg)