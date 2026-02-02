import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, SelectViewport } from '@/components/ui/select';
import { useTelemetryDevices, useTelemetryMetrics, useTelemetrySeries } from '@/features/telemetry/hooks';
import { LineChart } from '@/features/telemetry/LineChart';

export function TelemetryPage() {
  const { data: devices = [] } = useTelemetryDevices();
  const metrics = useTelemetryMetrics();
  const [deviceId, setDeviceId] = useState('');
  const [metricKey, setMetricKey] = useState(metrics[0]?.key ?? '');

  useEffect(() => {
    if (!deviceId && devices.length) {
      setDeviceId(devices[0].deviceId);
    }
  }, [deviceId, devices]);

  const { data: points = [], isLoading } = useTelemetrySeries(deviceId);
  const activeMetric = metrics.find((metric) => metric.key === metricKey);

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
          </div>
        </CardHeader>
        <CardContent>
          <div className="mb-4 text-sm text-muted-foreground">
            Metric: <span className="text-foreground">{activeMetric?.label}</span> ({activeMetric?.unit})
          </div>
          {isLoading ? <p className="text-sm text-muted-foreground">Loading telemetry...</p> : <LineChart points={points} />}
        </CardContent>
      </Card>
    </div>
  );
}
