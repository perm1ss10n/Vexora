# Testing Control Panel locally with Device Simulator (KONYX)

This guide runs a full local E2E loop without ESP32: MQTT broker + backend + simulated device + web Control Panel.

## 1) Start infrastructure (Mosquitto + InfluxDB)

From `backend/docker`:

```bash
docker compose up -d
```

Expected services:
- Mosquitto: `localhost:1883`
- InfluxDB: `localhost:8086`

## 2) Start backend (`:8080`)

From `backend`:

```bash
WEB_ALLOWED_ORIGIN=http://localhost:5173 \
MQTT_BROKER=tcp://localhost:1883 \
INFLUX_URL=http://localhost:8086 \
go run ./cmd/vexora-backend
```

Notes:
- If needed, also set `INFLUX_TOKEN`, `INFLUX_ORG`, `INFLUX_BUCKET` per your local setup.
- Default HTTP API is `http://localhost:8080`.

## 3) Start Device Simulator

From repository root:

```bash
go run ./tools/device-sim --device-id dev-123
```

Useful flags:
- `--mqtt-url tcp://localhost:1883`
- `--interval-ms 5000`
- `--fw fw-0.1.0`
- `--online=true`

Simulator behavior:
- Publishes retained state to `v1/dev/dev-123/state` on start and state changes.
- Publishes telemetry periodically to `v1/dev/dev-123/telemetry`.
- Listens on `v1/dev/dev-123/cmd`.
- Responds with ACK (`.../ack`) and events (`.../event`) for `ping`, `get_state`, `reboot`, `apply_cfg`.

## 4) Start web (`:5173`)

From `web`:

```bash
npm install
npm run dev
```

Ensure API base points to backend (`http://localhost:8080`) using existing web env/config.

## 5) Login / Register in Control Panel

1. Open `http://localhost:5173`.
2. Register a user (or login with an existing one).

## 6) Verify device + telemetry in UI

1. Go to **Devices** and confirm `dev-123` is visible and online.
2. Open device details and verify latest state/last seen are updating.
3. Go to **Telemetry** and verify live points appear on charts (`tempC`, `heatFlux`, `rssi`, `battV`).

## 7) Verify command flow (real ACK from simulator)

In **Commands** page for `dev-123`:

1. Run `ping` → expect ACK `ok=true`, event `PING`.
2. Run `get_state` → expect ACK `ok=true` and refreshed state.
3. Run `reboot` → expect ACK `ok=true`, event `REBOOT`, brief offline/online transition.
4. Run `apply_cfg` with telemetry config, for example:
   - `intervalMs=2000`
   - `minPublishMs=500`

Expected:
- ACK `ok=true` with updated config in ACK data.
- Event `CFG_APPLIED`.
- Telemetry cadence on charts changes within 1-2 intervals.

Validation rules for `apply_cfg` (simulator):
- `intervalMs`: `1000..3600000`
- `minPublishMs`: `0..3600000`
- `minPublishMs <= intervalMs` (when interval is provided)

On validation reject:
- ACK `ok=false` with code `CFG_REJECTED` or `BAD_CFG`.

## 8) Keep mock fallback untouched

Web mock/fallback paths are preserved; this flow validates the real API path with live MQTT ingestion.
