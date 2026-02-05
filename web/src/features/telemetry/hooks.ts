import { useQuery } from '@tanstack/react-query';
import { getDevices } from '@/api/devices';
import { getTelemetrySeries, TelemetrySeriesParams } from '@/api/telemetry';
import { Metric } from '@/api/types';
import { useAuthStore } from '@/store/auth';

export const telemetryMetrics: Metric[] = [
  { key: 'temp', unit: '°C', label: 'Temperature' },
  { key: 'voltage', unit: 'V', label: 'Voltage' },
];

export const useTelemetryDevices = () => {
  const accessToken = useAuthStore((state) => state.accessToken);
  const refreshAccessToken = useAuthStore((state) => state.refreshAccessToken);
  const logout = useAuthStore((state) => state.logout);

  return useQuery({
    queryKey: ['telemetry-devices'],
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

export const useTelemetryMetrics = () => telemetryMetrics;

export const useTelemetrySeries = (params: TelemetrySeriesParams | null) => {
  const accessToken = useAuthStore((state) => state.accessToken);
  const refreshAccessToken = useAuthStore((state) => state.refreshAccessToken);
  const logout = useAuthStore((state) => state.logout);

  return useQuery({
    queryKey: ['telemetry', params],
    queryFn: async () => {
      if (!accessToken || !params) {
        throw new Error('Missing telemetry query');
      }
      const { series } = await getTelemetrySeries(
        params,
        accessToken,
        async () => refreshAccessToken(),
        async () => logout()
      );
      return series;
    },
    enabled: !!accessToken && !!params,
  });
};
