import { apiRequestWithAuth } from './client';
import { TelemetrySeriesResponse } from './types';

export interface TelemetrySeriesParams {
  deviceId: string;
  metric: string;
  from: number;
  to: number;
  limit?: number;
}

export const getTelemetrySeries = async (
  params: TelemetrySeriesParams,
  accessToken: string,
  refreshToken?: () => Promise<string>,
  onUnauthorized?: () => Promise<void>
) => {
  const search = new URLSearchParams({
    metric: params.metric,
    from: String(params.from),
    to: String(params.to),
  });
  if (params.limit) {
    search.set('limit', String(params.limit));
  }
  const { data, accessToken: nextToken } = await apiRequestWithAuth<TelemetrySeriesResponse>(
    `/api/v1/devices/${encodeURIComponent(params.deviceId)}/telemetry?${search.toString()}`,
    accessToken,
    { method: 'GET' },
    refreshToken,
    onUnauthorized
  );
  return { series: data, accessToken: nextToken };
};
