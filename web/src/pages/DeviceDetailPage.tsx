import { useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useDeviceDetail } from '@/features/devices/hooks';
import { ApiError } from '@/api/client';

export function DeviceDetailPage() {
  const { deviceId = '' } = useParams();
  const { data, isLoading, isError, error } = useDeviceDetail(deviceId);

  const statusVariant = useMemo(() => {
    if (!data) return 'default';
    return data.device.status === 'online' ? 'success' : 'danger';
  }, [data]);

  if (isLoading) {
    return <p className="text-sm text-muted-foreground">Loading device...</p>;
  }

  if (isError) {
    if (error instanceof ApiError && error.status === 404) {
      return <p className="text-sm text-muted-foreground">Device not found.</p>;
    }
    return (
      <p className="text-sm text-destructive">
        {error instanceof Error ? error.message : 'Failed to load device.'}
      </p>
    );
  }

  if (!data) {
    return <p className="text-sm text-muted-foreground">No data.</p>;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold">{data.device.deviceId}</h2>
          <p className="text-sm text-muted-foreground">Device detail</p>
        </div>
        <Badge variant={statusVariant === 'default' ? 'accent' : (statusVariant as 'success' | 'danger')}>
          {data.device.status}
        </Badge>
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr,1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Device</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm text-muted-foreground">
              <div>
                Status: <span className="text-foreground">{data.device.status}</span>
              </div>
              <div>
                Firmware: <span className="text-foreground">{data.device.fwVersion ?? '—'}</span>
              </div>
              <div>
                Last seen:{' '}
                <span className="text-foreground">
                  {new Date(data.device.lastSeen * 1000).toLocaleString('ru-RU')}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>State</CardTitle>
          </CardHeader>
          <CardContent>
            {data.state ? (
              <div className="space-y-2 text-sm text-muted-foreground">
                <div>
                  Uptime: <span className="text-foreground">{data.state.uptime} sec</span>
                </div>
                <div>
                  Link: <span className="text-foreground">{data.state.link || '—'}</span>
                </div>
                <div>
                  IP: <span className="text-foreground">{data.state.ip || '—'}</span>
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No data.</p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Last telemetry</CardTitle>
        </CardHeader>
        <CardContent>
          {data.lastTelemetry && Object.keys(data.lastTelemetry.metrics).length > 0 ? (
            <div className="space-y-3 text-sm text-muted-foreground">
              <div>
                Timestamp:{' '}
                <span className="text-foreground">
                  {new Date(data.lastTelemetry.ts * 1000).toLocaleString('ru-RU')}
                </span>
              </div>
              <div className="grid gap-2 md:grid-cols-2">
                {Object.entries(data.lastTelemetry.metrics).map(([key, value]) => (
                  <div key={key} className="flex items-center justify-between rounded-md border border-border px-3 py-2">
                    <span className="text-foreground">{key}</span>
                    <span>{value}</span>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No data.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
