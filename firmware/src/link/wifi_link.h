#pragma once
#include "link.h"
#include <Arduino.h>
#include <WiFi.h>

struct WifiConfig {
  const char* ssid = nullptr;
  const char* pass = nullptr;
  uint32_t connectTimeoutMs = 15000;

  WifiConfig() = default;
  WifiConfig(const char* s, const char* p, uint32_t timeoutMs = 15000)
      : ssid(s), pass(p), connectTimeoutMs(timeoutMs) {}
};

class WifiLink : public ILink {
public:
  explicit WifiLink(WifiConfig cfg);

  LinkType type() const override { return LinkType::WIFI; }
  bool begin() override;
  bool isUp() const override;
  void loop() override;
  void disconnect() override;
  LinkStatus status() const override;

private:
  WifiConfig cfg_;
};