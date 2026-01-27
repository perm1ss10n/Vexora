# Vexora MQTT Protocol v1

## 0. Общие правила

### Версионирование
Все MQTT-топики начинаются с версии протокола:
- `v1/...`

### Идентификация
- `deviceId` — стабильный идентификатор устройства.
- Все команды и операции обязаны иметь `id` (UUID v4).

### Формат времени
- `ts` — Unix time в миллисекундах (`int64`).

### QoS / retained
- `telemetry`: QoS 0 (QoS 1 допускается для критичных каналов)
- `event`: QoS 1
- `state`: QoS 1, retained = true
- `cmd / cfg / ota`: QoS 1
- `ack`: QoS 1
- `cfg/status`: QoS 1, retained = true

### Безопасность
- Устройство подписывается только на свои топики
- Backend управляет ACL
- Устройства не публикуют сообщения вне своего namespace

---

## 1. Топики

### 1.1 Устройство → облако
- v1/dev/{deviceId}/telemetry
- v1/dev/{deviceId}/event
- v1/dev/{deviceId}/state (retained)

### 1.2 Облако → устройство
- v1/dev/{deviceId}/cmd
- v1/dev/{deviceId}/cfg
- v1/dev/{deviceId}/ota

### 1.3 Подтверждения
- v1/dev/{deviceId}/ack
- v1/dev/{deviceId}/cfg/status (retained)

---

## 2. Общий envelope сообщений

Каждое сообщение должно содержать:
- v — версия payload
- deviceId
- ts

---

## 3. Payload: Telemetry

Назначение: периодическая отправка измерений.

Топик:
- v1/dev/{deviceId}/telemetry

QoS: 0

---

## 4. Payload: Event

Назначение: ошибки, предупреждения, системные события.

Топик:
- v1/dev/{deviceId}/event

QoS: 1

---

## 5. Payload: State

Назначение: текущее состояние устройства (retained).

Топик:
- v1/dev/{deviceId}/state

QoS: 1, retained = true

---

## 6. Payload: Command (cmd)

Назначение: команды управления устройством.

Топик:
- v1/dev/{deviceId}/cmd

QoS: 1

---

## 7. Payload: Config (cfg)

Назначение: применение конфигурации устройства.

Топик:
- v1/dev/{deviceId}/cfg

QoS: 1

---

## 8. Payload: OTA

Назначение: OTA-обновление прошивки.

Топик:
- v1/dev/{deviceId}/ota

QoS: 1

---

## 9. Payload: ACK

Назначение: подтверждение выполнения команд.

Топик:
- v1/dev/{deviceId}/ack

QoS: 1

---

## 10. Payload: Config Status

Назначение: статус применения конфигурации (retained).

Топик:
- v1/dev/{deviceId}/cfg/status

QoS: 1, retained = true

---

## 11. Примечания

- MQTT-протокол является контрактом системы
- Изменения протокола возможны только через новую версию (v2)
