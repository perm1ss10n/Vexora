import { useEffect, useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, SelectViewport } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { useDevices } from '@/features/devices/hooks';
import { sendCommand } from '@/api/mock';
import { CommandRequest, CommandResult } from '@/api/types';

const commandOptions = [
  { label: 'Ping', value: 'ping' },
  { label: 'Reboot', value: 'reboot' },
  { label: 'Get state', value: 'get_state' },
  { label: 'Apply config', value: 'apply_cfg' },
] as const;

export function CommandsPage() {
  const { data: devices = [] } = useDevices();
  const [deviceId, setDeviceId] = useState('');
  const [commandType, setCommandType] = useState<CommandRequest['type']>('ping');
  const [intervalMs, setIntervalMs] = useState('5000');
  const [minPublishMs, setMinPublishMs] = useState('15000');
  const [log, setLog] = useState<CommandResult[]>([]);

  useEffect(() => {
    if (!deviceId && devices.length) {
      setDeviceId(devices[0].deviceId);
    }
  }, [deviceId, devices]);

  const needsConfig = commandType === 'apply_cfg';

  const payload = useMemo(() => {
    if (!needsConfig) return undefined;
    return {
      intervalMs: Number(intervalMs),
      minPublishMs: Number(minPublishMs),
    };
  }, [intervalMs, minPublishMs, needsConfig]);

  const handleSend = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!deviceId) return;

    const request: CommandRequest = {
      id: crypto.randomUUID(),
      deviceId,
      type: commandType,
      payload,
      createdAt: new Date().toISOString(),
    };

    const initial = await sendCommand(request);
    setLog((prev) => [initial, ...prev].slice(0, 8));

    setTimeout(() => {
      setLog((prev) =>
        prev.map((entry) =>
          entry.id === request.id
            ? { ...entry, status: Math.random() > 0.1 ? 'acked' : 'error' }
            : entry
        )
      );
    }, 1400);
  };

  return (
    <div className="grid gap-6 lg:grid-cols-[1.2fr,1fr]">
      <Card>
        <CardHeader>
          <CardTitle>Command dispatch</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={handleSend}>
            <div className="grid gap-3 md:grid-cols-2">
              <div>
                <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">Device</label>
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
              <div>
                <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">Command</label>
                <Select value={commandType} onValueChange={(value) => setCommandType(value as CommandRequest['type'])}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select command" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectViewport>
                      {commandOptions.map((command) => (
                        <SelectItem key={command.value} value={command.value}>
                          {command.label}
                        </SelectItem>
                      ))}
                    </SelectViewport>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {needsConfig ? (
              <div className="grid gap-3 md:grid-cols-2">
                <div>
                  <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">Interval (ms)</label>
                  <Input value={intervalMs} onChange={(event) => setIntervalMs(event.target.value)} />
                </div>
                <div>
                  <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">Min publish (ms)</label>
                  <Input value={minPublishMs} onChange={(event) => setMinPublishMs(event.target.value)} />
                </div>
              </div>
            ) : null}

            <Button type="submit">Send command</Button>
          </form>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Command log</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>Device</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {log.map((entry) => (
                <TableRow key={entry.id}>
                  <TableCell className="font-medium">{entry.type}</TableCell>
                  <TableCell className="text-muted-foreground">{entry.deviceId}</TableCell>
                  <TableCell>
                    <Badge
                      variant={
                        entry.status === 'acked'
                          ? 'success'
                          : entry.status === 'error'
                            ? 'danger'
                            : 'accent'
                      }
                    >
                      {entry.status}
                    </Badge>
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
