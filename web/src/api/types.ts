export type DeviceStatus = 'online' | 'offline';

export interface Device {
  deviceId: string;
  status: DeviceStatus;
  lastSeen: string;
  lastTelemetryAt?: string | null;
  fwVersion?: string | null;
}

export interface DeviceState {
  status: DeviceStatus;
  link: {
    type: 'lte' | 'wifi' | 'ethernet';
    rssi: number;
    ip: string;
  };
  fw: string;
  uptimeSec: number;
  ts: string;
}

export interface TelemetryPoint {
  ts: string;
  value: number;
}

export interface Metric {
  key: string;
  unit: string;
  label: string;
}

export interface CommandRequest {
  id: string;
  deviceId: string;
  type: 'ping' | 'reboot' | 'get_state' | 'apply_cfg';
  payload?: Record<string, string | number>;
  createdAt: string;
}

export interface CommandResult {
  id: string;
  deviceId: string;
  type: CommandRequest['type'];
  status: 'sent' | 'acked' | 'error';
  createdAt: string;
}
