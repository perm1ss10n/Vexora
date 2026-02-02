import { useQuery } from '@tanstack/react-query';
import { fetchDeviceTelemetry, fetchDevices, metrics } from '@/api/mock';

export const useTelemetryDevices = () =>
  useQuery({
    queryKey: ['telemetry-devices'],
    queryFn: fetchDevices,
  });

export const useTelemetryMetrics = () => metrics;

export const useTelemetrySeries = (deviceId: string) =>
  useQuery({
    queryKey: ['telemetry', deviceId],
    queryFn: () => fetchDeviceTelemetry(deviceId),
    enabled: !!deviceId,
  });
