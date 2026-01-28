#pragma once

#include <Arduino.h>
#include <stdint.h>
#include <stdio.h>

// Root: v1/dev/<deviceId>/...
static constexpr const char* VX_TOPIC_ROOT = "v1/dev/";

// Suffixes (must match backend topics.go)
static constexpr const char* VX_T_TELEMETRY   = "/telemetry";
static constexpr const char* VX_T_EVENT       = "/event";
static constexpr const char* VX_T_STATE       = "/state";
static constexpr const char* VX_T_ACK         = "/ack";
static constexpr const char* VX_T_CFG_STATUS  = "/cfg/status";
static constexpr const char* VX_T_LWT         = "/lwt";

// Device <- Cloud (commands/config/OTA)
// NOTE: backend topics.go currently doesn't list these yet, but firmware needs them for subscribe.
static constexpr const char* VX_T_CMD          = "/cmd";
static constexpr const char* VX_T_CFG          = "/cfg";
static constexpr const char* VX_T_OTA          = "/ota";

// Build "v1/dev/<deviceId><suffix>" into out.
// Returns true if ok and fits.
inline bool vx_build_topic(char* out, size_t outSize, const char* deviceId, const char* suffix) {
    if (!out || outSize == 0) return false;
    if (!deviceId || deviceId[0] == '\0') return false;
    if (!suffix || suffix[0] == '\0') return false;

    int n = snprintf(out, outSize, "%s%s%s", VX_TOPIC_ROOT, deviceId, suffix);
    return (n > 0) && ((size_t)n < outSize);
}