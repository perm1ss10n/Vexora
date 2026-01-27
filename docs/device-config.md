# Vexora Device Configuration v1

Цель: определить структуру конфигурации устройства, правила её применения,
версионирование и rollback.
Конфигурация является частью системного контракта и применяется удалённо через MQTT.

---

## 0) Общие принципы

- Конфигурация версионирована
- Применяется двухфазно
- Устройство всегда хранит:
  - active — текущая рабочая
  - previous — последняя рабочая
  - pending — временная (при применении)
- Некорректная конфигурация не может окирпичить устройство
- Любое применение подтверждается через MQTT

---

## 1) Жизненный цикл конфигурации

1. Backend отправляет cfg с cfgVersion
2. Устройство сохраняет конфигурацию как pending
3. Выполняется валидация
4. Конфигурация применяется
5. Проверяется восстановление связи (MQTT)
6. При успехе:
   - pending → active
   - старый active → previous
7. При ошибке:
   - rollback на previous
   - отправка cfg/status с ошибкой

---

## 2) Структура конфигурации (JSON)

{
  "version": 1,
  "network": {
    "wifi": {
      "enabled": true,
      "ssid": "",
      "password": ""
    },
    "gsm": {
      "enabled": true,
      "apn": "",
      "user": "",
      "password": ""
    }
  },
  "mqtt": {
    "broker": "mqtt.vexora.cloud",
    "port": 8883,
    "tls": true,
    "clientId": null
  },
  "telemetry": {
    "intervalSec": 30,
    "bufferEnabled": true,
    "bufferLimit": 1000
  },
  "device": {
    "name": "Vexora Box",
    "location": null
  }
}

---

## 3) Обязательные поля

- version — версия конфигурации
- network — минимум один канал связи должен быть enabled
- mqtt.broker — hostname или IP
- telemetry.intervalSec ≥ 1

---

## 4) Валидация конфигурации

Перед применением устройство обязано проверить:
- JSON корректен
- version > текущей active.version
- обязательные поля присутствуют

Network:
- хотя бы один из wifi.enabled или gsm.enabled = true
- если wifi.enabled=true → ssid не пустой
- если gsm.enabled=true → apn не пустой

Telemetry:
- intervalSec ∈ [1 … 86400]
- bufferLimit ≥ 0

---

## 5) Применение конфигурации

Порядок:
1. Сохранить pending
2. Применить network-настройки
3. Перезапустить сетевые интерфейсы
4. Подключиться к MQTT
5. При успехе → commit
6. При провале → rollback

---

## 6) Rollback

Rollback выполняется если:
- MQTT не восстановился за cfgApplyTimeoutSec
- произошла критическая ошибка сети
- конфигурация не прошла валидацию

---

## 7) Ограничения безопасности

- Конфигурация не может:
  - менять deviceId
  - менять provisioning credentials
  - отключить все каналы связи одновременно
- OTA и CFG не применяются параллельно

---

## 8) MQTT подтверждение

Топик:
- v1/dev/{deviceId}/cfg/status (retained)

Payload:
{
  "v": 1,
  "deviceId": "dev-123",
  "ts": 1730000000000,
  "activeVersion": 2,
  "pendingVersion": null,
  "lastApply": {
    "cfgVersion": 2,
    "ok": true,
    "error": null
  }
}

---

## 9) Версионирование

- Изменение структуры конфигурации → новая версия протокола
- Добавление полей допускается в рамках v1
- Устройство обязано игнорировать неизвестные поля
