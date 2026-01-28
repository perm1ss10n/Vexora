#include "wifi_link.h"

WifiLink::WifiLink(WifiConfig cfg) : cfg_(cfg) {}

bool WifiLink::begin() {
  if (!cfg_.ssid || cfg_.ssid[0] == '\0') return false;

  WiFi.mode(WIFI_STA);
  WiFi.setAutoReconnect(true);
  WiFi.begin(cfg_.ssid, cfg_.pass);

  const uint32_t start = millis();
  while (WiFi.status() != WL_CONNECTED && (millis() - start) < cfg_.connectTimeoutMs) {
    delay(100);
  }
  return WiFi.status() == WL_CONNECTED;
}

bool WifiLink::isUp() const {
  return WiFi.status() == WL_CONNECTED;
}

void WifiLink::loop() {
  // Для WiFi обычно достаточно autoReconnect. Тут можно добавить watchdog/логирование позже.
}

void WifiLink::disconnect() {
  WiFi.disconnect(true, true);
}

LinkStatus WifiLink::status() const {
  LinkStatus s;
  s.type = LinkType::WIFI;
  s.connected = isUp();
  if (s.connected) {
    s.rssi = WiFi.RSSI();
    static String ipStr;
    ipStr = WiFi.localIP().toString();
    s.ip = ipStr.c_str();
  }
  return s;
}