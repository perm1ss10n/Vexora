#pragma once
#include <stdint.h>

enum class LinkType : uint8_t { NONE = 0, WIFI = 1, GSM = 2 };

struct LinkStatus {
  LinkType type = LinkType::NONE;
  bool connected = false;
  int rssi = 0;          // если неизвестно — 0
  const char* ip = "";   // для WiFi, для GSM можно оставить ""
};

class ILink {
public:
  virtual ~ILink() = default;

  virtual LinkType type() const = 0;

  // Поднять физический канал (WiFi join / GSM attach)
  virtual bool begin() = 0;

  // “живой ли линк прямо сейчас”
  virtual bool isUp() const = 0;

  // Фоновое обслуживание (пинги/attach/пере-поднятие)
  virtual void loop() = 0;

  // Снять линк (disconnect)
  virtual void disconnect() = 0;

  // Диагностика
  virtual LinkStatus status() const = 0;
};