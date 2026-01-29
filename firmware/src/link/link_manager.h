#pragma once
#include "link.h"

class LinkManager {
public:
  LinkManager(ILink* wifi, ILink* gsm);

  bool begin();          // поднимаем предпочтительный линк
  void loop();           // поддержка + переключения
  ILink* active() const; // текущий линк
  LinkStatus status() const;

  // политика (можно вынести в config позже)
  void setPreferWifi(bool v) { preferWifi_ = v; }
  void setWifiProbeMs(uint32_t ms) { wifiProbeMs_ = ms; }

private:
  bool tryWifi_();
  bool tryGsm_();
  void switchTo(ILink* link);

  ILink* wifi_ = nullptr;
  ILink* gsm_ = nullptr;
  ILink* active_ = nullptr;

  bool preferWifi_ = true;
  uint32_t wifiProbeMs_ = 300000; // 5 минут
  uint32_t lastWifiProbeMs_ = 0;
};