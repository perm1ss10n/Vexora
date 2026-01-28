#include "link_manager.h"
#include <Arduino.h>

LinkManager::LinkManager(ILink* wifi, ILink* gsm) : wifi_(wifi), gsm_(gsm) {}

bool LinkManager::begin() {
  if (preferWifi_ && tryWifi_()) return true;
  if (tryGsm_()) return true;
  if (!preferWifi_ && tryWifi_()) return true;
  return false;
}

void LinkManager::loop() {
  if (active_) active_->loop();
  if (wifi_) wifi_->loop();
  if (gsm_) gsm_->loop();

  // Если активный линк упал — пробуем восстановить согласно приоритету
  if (active_ && !active_->isUp()) {
    active_->disconnect();
    active_ = nullptr;
    begin();
    return;
  }

  // Если сидим на GSM — периодически пробуем вернуть WiFi
  if (preferWifi_ && active_ && active_->type() == LinkType::GSM && wifi_) {
    const uint32_t now = millis();
    if (now - lastWifiProbeMs_ >= wifiProbeMs_) {
      lastWifiProbeMs_ = now;
      if (tryWifi_()) return;
    }
  }
}

ILink* LinkManager::active() const { return active_; }

LinkStatus LinkManager::status() const {
  if (active_) return active_->status();
  return LinkStatus{};
}

bool LinkManager::tryWifi_() {
  if (!wifi_) return false;
  if (wifi_->isUp()) { switchTo(wifi_); return true; }
  if (wifi_->begin()) { switchTo(wifi_); return true; }
  return false;
}

bool LinkManager::tryGsm_() {
  if (!gsm_) return false;
  if (gsm_->isUp()) { switchTo(gsm_); return true; }
  if (gsm_->begin()) { switchTo(gsm_); return true; }
  return false;
}

void LinkManager::switchTo(ILink* link) {
  if (active_ == link) return;
  if (active_) active_->disconnect();
  active_ = link;
}