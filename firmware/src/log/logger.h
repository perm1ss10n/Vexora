#pragma once
#include <Arduino.h>

namespace Logger
{
    void init(uint32_t baud = 115200);

    void i(const char *tag, const char *msg);
    void w(const char *tag, const char *msg);
    void e(const char *tag, const char *msg);

    void ifmt(const char *tag, const char *fmt, ...);
    void wfmt(const char *tag, const char *fmt, ...);
    void efmt(const char *tag, const char *fmt, ...);
}

// короткие макросы
#define LOGI(TAG, MSG) Logger::i((TAG), (MSG))
#define LOGW(TAG, MSG) Logger::w((TAG), (MSG))
#define LOGE(TAG, MSG) Logger::e((TAG), (MSG))

#define LOGIF(TAG, FMT, ...) Logger::ifmt((TAG), (FMT), ##__VA_ARGS__)
#define LOGWF(TAG, FMT, ...) Logger::wfmt((TAG), (FMT), ##__VA_ARGS__)
#define LOGEF(TAG, FMT, ...) Logger::efmt((TAG), (FMT), ##__VA_ARGS__)