export type DeviceStatus = 'online' | 'offline';

export interface Device {
  deviceId: string;
  status: DeviceStatus;
  lastSeen: number;
  fwVersion?: string | null;
}

export interface DeviceRuntimeState {
  uptime: number;
  link: string;
  ip: string;
}

export interface DeviceTelemetrySnapshot {
  ts: number;
  metrics: Record<string, number>;
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
