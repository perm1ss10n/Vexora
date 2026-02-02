import { useQuery } from '@tanstack/react-query';
import { fetchDeviceState, fetchLastTelemetry } from '@/api/mock';
import { getDevices } from '@/api/devices';
import { useAuthStore } from '@/store/auth';

export const useDevices = () => {
  const accessToken = useAuthStore((state) => state.accessToken);
  const refreshAccessToken = useAuthStore((state) => state.refreshAccessToken);
  const logout = useAuthStore((state) => state.logout);

  return useQuery({
    queryKey: ['devices'],
    queryFn: async () => {
      if (!accessToken) {
        throw new Error('No access token');
      }
      const { devices } = await getDevices(
        accessToken,
        async () => refreshAccessToken(),
        async () => logout()
      );
      return devices;
    },
    enabled: !!accessToken,
  });
};

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
