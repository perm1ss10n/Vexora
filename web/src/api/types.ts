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

export type CfgStatus = 'unknown' | 'pending' | 'applied' | 'rejected' | 'rolled_back';

export interface TelemetrySettings {
  intervalMs: number;
  minPublishMs: number;
}

export interface DeviceSettings {
  telemetry: TelemetrySettings;
  cfgStatus: CfgStatus;
}

export interface TelemetryPoint {
  ts: number;
  value: number;
}

export interface TelemetrySeriesResponse {
  deviceId: string;
  metric: string;
  from: number;
  to: number;
  points: TelemetryPoint[];
}

export interface Metric {
  key: string;
  unit: string;
  label: string;
}

export type CommandType = 'ping' | 'reboot' | 'get_state' | 'apply_cfg';

export interface CommandAck {
  v: number;
  id: string;
  deviceId: string;
  ts: number;
  ok: boolean;
  code?: string;
  msg?: string;
  data?: Record<string, unknown>;
}
