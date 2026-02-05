import { useEffect, useMemo, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, SelectViewport } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { ApiError } from '@/api/client';
import { TelemetrySeriesParams } from '@/api/telemetry';
import { useTelemetryDevices, useTelemetryMetrics, useTelemetrySeries } from '@/features/telemetry/hooks';
import { LineChart } from '@/features/telemetry/LineChart';

const ranges = [
  { key: '15m', label: '15 min', seconds: 15 * 60 },
  { key: '1h', label: '1 hour', seconds: 60 * 60 },
  { key: '24h', label: '24 hours', seconds: 24 * 60 * 60 },
];

export function TelemetryPage() {
  const { data: devices = [] } = useTelemetryDevices();
  const metrics = useTelemetryMetrics();
  const [deviceId, setDeviceId] = useState('');
  const [metricKey, setMetricKey] = useState(metrics[0]?.key ?? '');
  const [rangeKey, setRangeKey] = useState(ranges[0].key);
  const [queryParams, setQueryParams] = useState<TelemetrySeriesParams | null>(null);

  useEffect(() => {
    if (!deviceId && devices.length) {
      setDeviceId(devices[0].deviceId);
    }
  }, [deviceId, devices]);

  const activeRange = useMemo(() => ranges.find((range) => range.key === rangeKey) ?? ranges[0], [rangeKey]);
  const { data: series, isLoading, error } = useTelemetrySeries(queryParams);
  const activeMetric = metrics.find((metric) => metric.key === metricKey);
  const points = series?.points ?? [];

  const errorMessage = useMemo(() => {
    if (!error) {
      return '';
    }
    if (error instanceof ApiError) {
      if (error.status === 501) {
        return 'Influx disabled on backend';
      }
      return error.message || `Request failed (${error.status})`;
    }
    return (error as Error).message || 'Failed to load telemetry';
  }, [error]);

  const handleLoad = () => {
    if (!deviceId || !metricKey) {
      return;
    }
    const now = Math.floor(Date.now() / 1000);
    const from = now - activeRange.seconds;
    setQueryParams({
      deviceId,
      metric: metricKey,
      from,
      to: now,
      limit: 2000,
    });
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <CardTitle>Telemetry charts</CardTitle>
          <div className="flex flex-wrap gap-3">
            <div className="w-56">
              <Select value={deviceId} onValueChange={setDeviceId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select device" />
                </SelectTrigger>
                <SelectContent>
                  <SelectViewport>
                    {devices.map((device) => (
                      <SelectItem key={device.deviceId} value={device.deviceId}>
                        {device.deviceId}
                      </SelectItem>
                    ))}
                  </SelectViewport>
                </SelectContent>
              </Select>
            </div>
            <div className="w-48">
              <Select value={metricKey} onValueChange={setMetricKey}>
                <SelectTrigger>
                  <SelectValue placeholder="Select metric" />
                </SelectTrigger>
                <SelectContent>
                  <SelectViewport>
                    {metrics.map((metric) => (
                      <SelectItem key={metric.key} value={metric.key}>
                        {metric.label}
                      </SelectItem>
                    ))}
                  </SelectViewport>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-wrap gap-2">
              {ranges.map((range) => (
                <Button
                  key={range.key}
                  type="button"
                  size="sm"
                  variant={range.key === rangeKey ? 'default' : 'outline'}
                  onClick={() => setRangeKey(range.key)}
                >
                  {range.label}
                </Button>
              ))}
            </div>
            <Button type="button" onClick={handleLoad} disabled={!deviceId || !metricKey}>
              Load
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="mb-4 text-sm text-muted-foreground">
            Metric: <span className="text-foreground">{activeMetric?.label}</span> ({activeMetric?.unit})
          </div>
          {errorMessage ? (
            <p className="text-sm text-destructive">{errorMessage}</p>
          ) : isLoading ? (
            <p className="text-sm text-muted-foreground">Loading telemetry...</p>
          ) : queryParams ? (
            <LineChart points={points} />
          ) : (
            <p className="text-sm text-muted-foreground">Choose parameters and click Load.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
