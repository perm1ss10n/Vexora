# Device Simulator (KONYX)

Local MQTT device simulator for end-to-end Control Panel testing without ESP32.

## Run

From repository root:

```bash
go run ./tools/device-sim --device-id dev-123
```

### Flags

- `--device-id` (default: `dev-123`)
- `--mqtt-url` (default: `$MQTT_URL` or `tcp://localhost:1883`)
- `--interval-ms` (default: `5000`)
- `--fw` (default: `fw-0.1.0`)
- `--online` (default: `true`)

## Topics

- publish retained state: `v1/dev/<deviceId>/state`
- publish telemetry: `v1/dev/<deviceId>/telemetry`
- subscribe commands: `v1/dev/<deviceId>/cmd`
- publish ack: `v1/dev/<deviceId>/ack`
- publish events: `v1/dev/<deviceId>/event`

## Commands

- `ping`
- `get_state`
- `reboot`
- `apply_cfg` (`cfg.telemetry.intervalMs`, `cfg.telemetry.minPublishMs`)
