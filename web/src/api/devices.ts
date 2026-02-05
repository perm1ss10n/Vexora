import { apiRequestWithAuth } from './client';
import { Device, DeviceRuntimeState, DeviceSettings, DeviceTelemetrySnapshot } from './types';

export const getDevices = async (
  accessToken: string,
  refreshToken?: () => Promise<string>,
  onUnauthorized?: () => Promise<void>
) => {
  const { data, accessToken: nextToken } = await apiRequestWithAuth<Device[]>(
    '/api/v1/devices',
    accessToken,
    { method: 'GET' },
    refreshToken,
    onUnauthorized
  );
  return { devices: data ?? [], accessToken: nextToken };
};

export interface DeviceDetailResponse {
  device: Device;
  state: DeviceRuntimeState | null;
  lastTelemetry: DeviceTelemetrySnapshot | null;
  settings: DeviceSettings;
  settingsSource?: string;
}

export const getDeviceDetail = async (
  deviceId: string,
  accessToken: string,
  refreshToken?: () => Promise<string>,
  onUnauthorized?: () => Promise<void>
) => {
  const { data, accessToken: nextToken } = await apiRequestWithAuth<DeviceDetailResponse>(
    `/api/v1/devices/${deviceId}`,
    accessToken,
    { method: 'GET' },
    refreshToken,
    onUnauthorized
  );
  return { detail: data, accessToken: nextToken };
};
