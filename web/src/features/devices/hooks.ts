import { useQuery } from '@tanstack/react-query';
import { getDeviceDetail, getDevices } from '@/api/devices';
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

export const useDeviceDetail = (deviceId: string) => {
  const accessToken = useAuthStore((state) => state.accessToken);
  const refreshAccessToken = useAuthStore((state) => state.refreshAccessToken);
  const logout = useAuthStore((state) => state.logout);

  return useQuery({
    queryKey: ['device-detail', deviceId],
    queryFn: async () => {
      if (!accessToken) {
        throw new Error('No access token');
      }
      const { detail } = await getDeviceDetail(
        deviceId,
        accessToken,
        async () => refreshAccessToken(),
        async () => logout()
      );
      return detail;
    },
    enabled: !!deviceId && !!accessToken,
  });
};
