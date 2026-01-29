#pragma once
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>   // snprintf

enum VxTopicKind : uint8_t {
    VX_T_TELEMETRY,
    VX_T_EVENT,
    VX_T_STATE,
    VX_T_ACK,
    VX_T_CFG_STATUS,
    VX_T_LWT,
    VX_T_CMD,
    VX_T_CFG,
    VX_T_OTA
};

inline const char* vx_topic_suffix(VxTopicKind k) {
    switch (k) {
        case VX_T_TELEMETRY:  return "telemetry";
        case VX_T_EVENT:      return "event";
        case VX_T_STATE:      return "state";
        case VX_T_ACK:        return "ack";
        case VX_T_CFG_STATUS: return "cfg/status";
        case VX_T_LWT:        return "lwt";
        case VX_T_CMD:        return "cmd";
        case VX_T_CFG:        return "cfg";
        case VX_T_OTA:        return "ota";
        default:              return "";
    }
}

// v1/dev/<deviceId>/<suffix>
inline bool vx_build_topic(char* out, size_t outSize, const char* deviceId, VxTopicKind kind) {
    if (!out || outSize == 0 || !deviceId || deviceId[0] == '\0') return false;

    const char* suf = vx_topic_suffix(kind);
    if (!suf || suf[0] == '\0') return false;

    int n = snprintf(out, outSize, "v1/dev/%s/%s", deviceId, suf);
    return n > 0 && (size_t)n < outSize;
}