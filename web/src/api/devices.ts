import { apiRequestWithAuth } from './client';
import { Device } from './types';

interface DevicesResponse {
  devices: Device[];
}

export const getDevices = async (
  accessToken: string,
  refreshToken?: () => Promise<string>,
  onUnauthorized?: () => Promise<void>
) => {
  const { data, accessToken: nextToken } = await apiRequestWithAuth<DevicesResponse>(
    '/api/v1/devices',
    accessToken,
    { method: 'GET' },
    refreshToken,
    onUnauthorized
  );
  return { devices: data.devices ?? [], accessToken: nextToken };
};
