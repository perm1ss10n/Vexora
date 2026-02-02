import { useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useDeviceState, useLastTelemetry } from '@/features/devices/hooks';

const commandButtons = [
  { label: 'Ping', value: 'ping' },
  { label: 'Get state', value: 'get_state' },
  { label: 'Reboot', value: 'reboot' },
];

export function DeviceDetailPage() {
  const { deviceId = '' } = useParams();
  const { data: state } = useDeviceState(deviceId);
  const { data: telemetry = [] } = useLastTelemetry(deviceId);
  const [lastCommand, setLastCommand] = useState('');

  const statusVariant = useMemo(() => {
    if (!state) return 'default';
    return state.status === 'online' ? 'success' : 'danger';
  }, [state]);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold">{deviceId}</h2>
          <p className="text-sm text-muted-foreground">Premium device overview</p>
        </div>
        <Badge variant={statusVariant === 'default' ? 'accent' : (statusVariant as 'success' | 'danger')}>
          {state?.status ?? 'loading'}
        </Badge>
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr,1fr]">
        <Card>
          <CardHeader>
            <CardTitle>State (retained)</CardTitle>
          </CardHeader>
          <CardContent>
            {state ? (
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2 text-sm text-muted-foreground">
                  <div>Status: <span className="text-foreground">{state.status}</span></div>
                  <div>Link type: <span className="text-foreground">{state.link.type}</span></div>
                  <div>RSSI: <span className="text-foreground">{state.link.rssi} dBm</span></div>
                  <div>IP: <span className="text-foreground">{state.link.ip}</span></div>
                </div>
                <div className="space-y-2 text-sm text-muted-foreground">
                  <div>FW: <span className="text-foreground">{state.fw}</span></div>
                  <div>Uptime: <span className="text-foreground">{state.uptimeSec} sec</span></div>
                  <div>Timestamp: <span className="text-foreground">{new Date(state.ts).toLocaleString('ru-RU')}</span></div>
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">Loading state...</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Quick commands</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {commandButtons.map((command) => (
              <Button
                key={command.value}
                variant="outline"
                className="w-full justify-center"
                onClick={() => setLastCommand(command.label)}
              >
                {command.label}
              </Button>
            ))}
            {lastCommand ? (
              <p className="text-xs text-muted-foreground">Last issued: {lastCommand}</p>
            ) : null}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Last telemetry</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Metric</TableHead>
                <TableHead>Value</TableHead>
                <TableHead>Unit</TableHead>
                <TableHead>Timestamp</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {telemetry.map((metric) => (
                <TableRow key={metric.key}>
                  <TableCell className="font-medium text-foreground">{metric.key}</TableCell>
                  <TableCell>{metric.value}</TableCell>
                  <TableCell className="text-muted-foreground">{metric.unit}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {new Date(metric.ts).toLocaleString('ru-RU')}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
