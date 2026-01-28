#include "log/logger.h"

void Logger::init() {
    // Ничего особенного пока не надо: Serial уже поднят в main.cpp
}

static void logLine(const char* level, const char* tag, const char* msg) {
    Serial.print('[');
    Serial.print(level);
    Serial.print("] ");
    Serial.print(tag);
    Serial.print(": ");
    Serial.println(msg);
}

void Logger::info(const char* tag, const char* msg) {
    logLine("INFO", tag, msg);
}

void Logger::warn(const char* tag, const char* msg) {
    logLine("WARN", tag, msg);
}

void Logger::error(const char* tag, const char* msg) {
    logLine("ERR", tag, msg);
}