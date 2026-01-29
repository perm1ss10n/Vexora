#include "log/logger.h"

#include <Arduino.h>
#include <cstdarg>
#include <cstdio>

static void vlog(const char *level, const char *tag, const char *fmt, va_list ap)
{
    char msg[256];
    vsnprintf(msg, sizeof(msg), fmt ? fmt : "", ap);
    Serial.printf("[%s][%s] %s\n", level, tag ? tag : "?", msg);
}

namespace Logger
{

    void init(uint32_t baud)
    {
        Serial.begin(baud);
        delay(50);
    }

    void i(const char *tag, const char *msg) { Serial.printf("[I][%s] %s\n", tag ? tag : "?", msg ? msg : ""); }
    void w(const char *tag, const char *msg) { Serial.printf("[W][%s] %s\n", tag ? tag : "?", msg ? msg : ""); }
    void e(const char *tag, const char *msg) { Serial.printf("[E][%s] %s\n", tag ? tag : "?", msg ? msg : ""); }

    void ifmt(const char *tag, const char *fmt, ...)
    {
        va_list ap;
        va_start(ap, fmt);
        vlog("I", tag, fmt, ap);
        va_end(ap);
    }

    void wfmt(const char *tag, const char *fmt, ...)
    {
        va_list ap;
        va_start(ap, fmt);
        vlog("W", tag, fmt, ap);
        va_end(ap);
    }

    void efmt(const char *tag, const char *fmt, ...)
    {
        va_list ap;
        va_start(ap, fmt);
        vlog("E", tag, fmt, ap);
        va_end(ap);
    }

} // namespace Logger