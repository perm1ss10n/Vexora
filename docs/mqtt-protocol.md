# Vexora MQTT Protocol v1

## Версионирование
Все топики используют версию протокола:
v1/

## Топики устройства → облако
- v1/dev/{deviceId}/telemetry
- v1/dev/{deviceId}/event
- v1/dev/{deviceId}/state   (retained)

## Топики облако → устройство
- v1/dev/{deviceId}/cmd
- v1/dev/{deviceId}/cfg
- v1/dev/{deviceId}/ota

## Подтверждения
- v1/dev/{deviceId}/ack
- v1/dev/{deviceId}/cfg/status   (retained)

## Общие правила
- QoS 1 для команд, конфигов и OTA
- Все команды имеют уникальный `id`
- Устройство обязано отвечать ACK
- Устройство подписывается **только на свои топики**
- Backend контролирует ACL