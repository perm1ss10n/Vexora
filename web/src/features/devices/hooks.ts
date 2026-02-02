import { useQuery } from '@tanstack/react-query';
import { fetchDeviceState, fetchDevices, fetchLastTelemetry } from '@/api/mock';

export const useDevices = () =>
  useQuery({
    queryKey: ['devices'],
    queryFn: fetchDevices,
  });

export const useDeviceState = (deviceId: string) =>
  useQuery({
    queryKey: ['device-state', deviceId],
    queryFn: () => fetchDeviceState(deviceId),
    enabled: !!deviceId,
  });

export const useLastTelemetry = (deviceId: string) =>
  useQuery({
    queryKey: ['device-telemetry-last', deviceId],
    queryFn: () => fetchLastTelemetry(deviceId),
    enabled: !!deviceId,
  });
