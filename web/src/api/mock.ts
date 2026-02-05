import { CommandRequest, CommandResult, Device, DeviceRuntimeState, Metric, TelemetryPoint } from './types';

const deviceIds = [
  'KNY-AX91-001',
  'KNY-AX91-002',
  'KNY-AX91-003',
  'KNY-AX92-017',
  'KNY-AX92-018',
  'KNY-ZT10-404',
  'KNY-ZT10-405',
  'KNY-ZT11-909',
];

const now = () => new Date();

const randomStatus = (index: number) => (index % 3 === 0 ? 'offline' : 'online');

export const mockDevices: Device[] = deviceIds.map((deviceId, index) => {
  const lastSeen = new Date(now().getTime() - (index + 1) * 1000 * 60 * 5);
  return {
    deviceId,
    status: randomStatus(index),
    lastSeen: Math.floor(lastSeen.getTime() / 1000),
    fwVersion: `v2.4.${index}`,
  };
});

export const fetchDevices = async (): Promise<Device[]> => {
  await new Promise((resolve) => setTimeout(resolve, 250));
  return mockDevices;
};

export const metrics: Metric[] = [
  { key: 'temperature', unit: '°C', label: 'Temperature' },
  { key: 'voltage', unit: 'V', label: 'Voltage' },
  { key: 'humidity', unit: '%', label: 'Humidity' },
  { key: 'pressure', unit: 'kPa', label: 'Pressure' },
];

export const fetchDeviceState = async (deviceId: string): Promise<DeviceRuntimeState> => {
  await new Promise((resolve) => setTimeout(resolve, 200));
  const deviceIndex = deviceIds.indexOf(deviceId);
  return {
    uptime: 100230 + deviceIndex * 543,
    link: deviceIndex % 2 === 0 ? 'lte' : 'wifi',
    ip: `10.21.${deviceIndex + 10}.45`,
  };
};

export const fetchDeviceTelemetry = async (deviceId: string): Promise<TelemetryPoint[]> => {
  await new Promise((resolve) => setTimeout(resolve, 200));
  const base = now().getTime();
  return Array.from({ length: 60 }).map((_, index) => {
    const ts = new Date(base - (59 - index) * 60 * 1000);
    return {
      ts: ts.toISOString(),
      value: 50 + Math.sin(index / 5) * 8 + (deviceId.length % 4) * 2 + Math.random() * 2,
    };
  });
};

export const fetchLastTelemetry = async (deviceId: string) => {
  await new Promise((resolve) => setTimeout(resolve, 150));
  const base = now().getTime();
  return metrics.map((metric, index) => ({
    key: metric.key,
    unit: metric.unit,
    value: Number((Math.random() * 40 + 20).toFixed(1)),
    ts: new Date(base - index * 1000 * 60 * 2).toISOString(),
  }));
};

export const sendCommand = async (request: CommandRequest): Promise<CommandResult> => {
  await new Promise((resolve) => setTimeout(resolve, 300));
  return {
    id: request.id,
    deviceId: request.deviceId,
    type: request.type,
    status: 'sent',
    createdAt: request.createdAt,
  };
};
