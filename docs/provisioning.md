# Vexora Provisioning v1

Цель: определить жизненный цикл устройства, генерацию `deviceId`, первичную активацию,
хранение токенов/ключей и сценарий первого запуска без интернета.

Provisioning — часть контракта системы. Firmware, Backend и App должны реализовать это одинаково.

---

## 0) Термины

- **Factory** — устройство в заводском состоянии, без привязки к аккаунту/тенанту.
- **Provisioned** — устройство получило первичный конфиг и креды для связи.
- **Activated** — устройство привязано к клиенту (tenant) и разрешено в прод.
- **Revoked** — устройство отозвано (токены инвалидированы), доступ запрещён.
- **Bootstrap** — минимальная конфигурация, которая позволяет устройству впервые выйти в сеть.

---

## 1) Жизненный цикл устройства

Состояния:

1) **FACTORY**
- Нет `deviceToken`
- Может иметь `deviceId` (зашит/сгенерен) или временный `hwId`
- Входит в режим provisioning (AP/BLE) при первом запуске или по кнопке

2) **PROVISIONED**
- Сохранён bootstrap config
- Сохранены креды для MQTT (deviceToken/cert)
- Устройство может подключаться к брокеру, публиковать `state`, принимать `cfg`

3) **ACTIVATED**
- Backend разрешает полноценную работу (ACL, доступ к топикам, команды, OTA)
- Device registry содержит tenantId и политику доступа

4) **REVOKED**
- Доступ блокируется (ACL deny), токен/сертификат считается недействительным
- Устройство обязано перейти в OFFLINE/PROVISIONING (в зависимости от политики)

---

## 2) Идентификаторы: deviceId и hwId

### 2.1 deviceId
`deviceId` — стабильный идентификатор устройства в системе.

Требования:
- уникален
- не меняется в течение жизни устройства
- не зависит от Wi-Fi/GSM настроек

Рекомендуемый формат (MVP):
- `dev-` + 12–16 hex символов (например, `dev-a1b2c3d4e5f6`)

### 2.2 hwId (опционально)
`hwId` — аппаратный идентификатор (MAC, chip id, серийник).
Используется только как источник энтропии/проверки, но не публикуется в открытом виде.

---

## 3) Генерация deviceId

Выбор для MVP (простое и надёжное):

- deviceId генерируется **на устройстве** при первом запуске
- источник: ESP32 chip id / MAC (хэшируется)
- сохраняется в NVS/flash как immutable

Псевдологика:
- read chip id
- deviceId = `dev-` + hex(sha256(chipId))[0:12]
- persist deviceId

Важно:
- конфигурация **не может менять** deviceId (см. device-config.md)

---

## 4) Первичная активация (Activation)

Provisioning делим на 2 уровня:

### 4.1 Bootstrap provisioning (локально, без интернета)
Назначение: дать устройству возможность подключиться к MQTT и получить дальнейшие настройки.

Каналы:
- локальный Wi-Fi AP (рекомендуется для MVP)
- BLE (опционально)

Что передаёт Vexora App:
- Wi-Fi credentials (optional)
- GSM APN credentials (optional)
- MQTT broker host/port/tls
- `activationCode` (обязательно)
- (опционально) human-friendly name/location

Результат:
- устройство получает bootstrap config и сохраняет как `active` v1
- устройство пытается подключиться и выйти в PROVISIONED

### 4.2 Cloud activation (через backend)
Назначение: привязать устройство к tenantId и выдать реальные креды/ACL.

Сценарий:
1) Пользователь в Vexora App создаёт устройство в Cloud → получает `activationCode`
2) Вводит `activationCode` в локальном provisioning
3) Устройство подключается к MQTT и публикует `state`/`event` (PROVISIONED)
4) Backend подтверждает активацию и выдаёт:
   - `deviceToken` (или короткоживущий token + refresh)
   - ACL правила
   - базовый `cfg` (telemetry interval, buffer policy, etc.)
5) Устройство сохраняет токен/креды и переходит в ACTIVATED

---

## 5) Activation Code

Требования:
- одноразовый
- ограниченный по времени (например 10–30 минут)
- связывает устройство и tenant

Формат (MVP):
- 8–12 символов base32/hex (например, `K7P3-9Q2F`)

Правило:
- activationCode нельзя использовать повторно после успешной привязки.

---

## 6) Хранение токенов/ключей на устройстве

Минимально для MVP:

- `deviceId` — NVS, immutable
- `deviceToken` — NVS, обновляемый
- `bootstrap cfg` / `active cfg` / `previous cfg` — flash/NVS
- последний статус активации — NVS (`activated=true/false`)

Защита:
- не логировать токены
- не отправлять токены в telemetry/state
- при wipe/reset — токен удаляется, deviceId остаётся (по политике)

---

## 7) Первый запуск без интернета (обязательный сценарий)

Сценарий:

1) Устройство включается → нет конфига → входит в PROVISIONING
2) Поднимает локальный AP `VEXORA-SETUP-XXXX`
3) Vexora App подключается к AP и открывает мастер настройки
4) Пользователь вводит хотя бы один канал связи:
   - Wi-Fi (ssid/pass) **или** GSM (apn)
5) Пользователь вводит `activationCode` (полученный заранее)
6) Устройство сохраняет bootstrap cfg и пытается подключиться:
   - если получилось → PROVISIONED → ACTIVATED
   - если не получилось → остаётся в PROVISIONING/OFFLINE с retry

Примечание:
- Если интернет появится позже (например, Wi-Fi появится), устройство должно само активироваться.

---

## 8) Сброс устройства (Factory reset)

Политика для MVP:

- кнопка/команда `factoryReset`
- очищает:
  - deviceToken
  - active/previous cfg
  - provisioning flags
- оставляет:
  - deviceId (по умолчанию)

После reset:
- устройство возвращается в FACTORY и снова требует provisioning

---

## 9) MQTT следы provisioning (минимум)

На этапе provisioning устройство использует те же топики, но backend может ограничивать функциональность:

- `state` публикуется всегда
- `telemetry` может быть запрещена до ACTIVATED
- команды допускаются только базовые (`ping`, `requestState`, `factoryReset`)

Backend должен уметь отличать PROVISIONED vs ACTIVATED по registry.

---

## 10) Что реализуем в MVP (чек-лист)

- локальный AP provisioning
- activationCode от backend
- генерация deviceId на устройстве
- хранение deviceToken
- factory reset
