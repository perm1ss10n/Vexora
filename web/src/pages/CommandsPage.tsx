import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, SelectViewport } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { ApiError } from '@/api/client';
import { sendCommand } from '@/api/commands';
import { CommandAck, CommandType } from '@/api/types';
import { useDevices } from '@/features/devices/hooks';
import { useAuthStore } from '@/store/auth';

type CommandState = {
  ack?: CommandAck;
  error?: string;
  isLoading: boolean;
};

const createCommandState = (): CommandState => ({ isLoading: false });

export function CommandsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const accessToken = useAuthStore((state) => state.accessToken);
  const refreshAccessToken = useAuthStore((state) => state.refreshAccessToken);
  const logout = useAuthStore((state) => state.logout);
  const { data: devices = [] } = useDevices();
  const [deviceId, setDeviceId] = useState('');
  const [intervalMs, setIntervalMs] = useState('5000');
  const [minPublishMs, setMinPublishMs] = useState('');
  const [pingState, setPingState] = useState<CommandState>(createCommandState());
  const [stateState, setStateState] = useState<CommandState>(createCommandState());
  const [rebootState, setRebootState] = useState<CommandState>(createCommandState());
  const [applyState, setApplyState] = useState<CommandState>(createCommandState());

  useEffect(() => {
    if (!deviceId && devices.length) {
      setDeviceId(devices[0].deviceId);
    }
  }, [deviceId, devices]);

  const commandErrorMessage = useMemo(() => {
    return (error: unknown) => {
      if (error instanceof ApiError) {
        return error.message || `Request failed (${error.status})`;
      }
      return (error as Error).message || 'Request failed';
    };
  }, []);

  const handleUnauthorized = async () => {
    await logout();
    navigate('/login');
  };

  const runCommand = async (
    type: CommandType,
    setState: (value: CommandState) => void,
    params?: Record<string, unknown>
  ) => {
    if (!deviceId || !accessToken) {
      return;
    }
    setState({ isLoading: true });
    try {
      const { ack } = await sendCommand(
        deviceId,
        type,
        accessToken,
        params,
        undefined,
        async () => refreshAccessToken(),
        handleUnauthorized
      );
      setState({ isLoading: false, ack });
      if (type === 'get_state' && ack.ok) {
        await queryClient.invalidateQueries({ queryKey: ['device-detail', deviceId] });
      }
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        await handleUnauthorized();
      }
      setState({ isLoading: false, error: commandErrorMessage(error) });
    }
  };

  const handleApplyConfig = async (event: React.FormEvent) => {
    event.preventDefault();
    const intervalValue = Number(intervalMs);
    if (Number.isNaN(intervalValue) || intervalValue < 1000 || intervalValue > 3_600_000) {
      setApplyState({ isLoading: false, error: 'Interval must be between 1000 and 3600000 ms.' });
      return;
    }
    const minPublishValue = minPublishMs.trim() === '' ? undefined : Number(minPublishMs);
    if (
      typeof minPublishValue === 'number' &&
      (Number.isNaN(minPublishValue) || minPublishValue < 0 || minPublishValue > 3_600_000)
    ) {
      setApplyState({ isLoading: false, error: 'Min publish must be between 0 and 3600000 ms.' });
      return;
    }

    const telemetry: Record<string, number> = { intervalMs: intervalValue };
    if (typeof minPublishValue === 'number') {
      telemetry.minPublishMs = minPublishValue;
    }

    await runCommand('apply_cfg', setApplyState, {
      cfg: {
        telemetry,
      },
    });
  };

  const handleReboot = async () => {
    if (!window.confirm('Reboot device now?')) {
      return;
    }
    await runCommand('reboot', setRebootState);
  };

  const renderAck = (ack?: CommandAck) => {
    if (!ack) {
      return <p className="text-sm text-muted-foreground">No ACK yet.</p>;
    }
    const timestamp = new Date(ack.ts).toLocaleString();
    return (
      <dl className="mt-3 grid gap-2 text-sm">
        <div className="flex items-center justify-between">
          <dt className="text-muted-foreground">Status</dt>
          <dd>
            <Badge variant={ack.ok ? 'success' : 'danger'}>{ack.ok ? 'OK' : 'FAIL'}</Badge>
          </dd>
        </div>
        <div className="flex items-center justify-between">
          <dt className="text-muted-foreground">Code</dt>
          <dd className="font-medium text-foreground">{ack.code || '—'}</dd>
        </div>
        <div className="flex items-center justify-between">
          <dt className="text-muted-foreground">Message</dt>
          <dd className="text-right text-foreground">{ack.msg || '—'}</dd>
        </div>
        <div className="flex items-center justify-between">
          <dt className="text-muted-foreground">Timestamp</dt>
          <dd className="text-right text-foreground">{timestamp}</dd>
        </div>
      </dl>
    );
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <CardTitle>Commands</CardTitle>
          <div className="w-64">
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
        </CardHeader>
        <CardContent>
          {!deviceId && <p className="text-sm text-muted-foreground">Select a device to send commands.</p>}
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Ping</CardTitle>
          </CardHeader>
          <CardContent>
            <Button
              type="button"
              onClick={() => runCommand('ping', setPingState)}
              disabled={!deviceId || pingState.isLoading}
            >
              {pingState.isLoading ? 'Sending...' : 'Send ping'}
            </Button>
            {pingState.error && <p className="mt-3 text-sm text-destructive">{pingState.error}</p>}
            {renderAck(pingState.ack)}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Get state</CardTitle>
          </CardHeader>
          <CardContent>
            <Button
              type="button"
              onClick={() => runCommand('get_state', setStateState)}
              disabled={!deviceId || stateState.isLoading}
            >
              {stateState.isLoading ? 'Sending...' : 'Request state'}
            </Button>
            {stateState.error && <p className="mt-3 text-sm text-destructive">{stateState.error}</p>}
            {renderAck(stateState.ack)}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Reboot</CardTitle>
          </CardHeader>
          <CardContent>
            <Button
              type="button"
              variant="destructive"
              onClick={handleReboot}
              disabled={!deviceId || rebootState.isLoading}
            >
              {rebootState.isLoading ? 'Sending...' : 'Reboot device'}
            </Button>
            {rebootState.error && <p className="mt-3 text-sm text-destructive">{rebootState.error}</p>}
            {renderAck(rebootState.ack)}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Apply config</CardTitle>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={handleApplyConfig}>
              <div className="grid gap-3 md:grid-cols-2">
                <div>
                  <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">
                    Interval (ms)
                  </label>
                  <Input
                    type="number"
                    min={1000}
                    max={3600000}
                    value={intervalMs}
                    onChange={(event) => setIntervalMs(event.target.value)}
                  />
                </div>
                <div>
                  <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">
                    Min publish (ms)
                  </label>
                  <Input
                    type="number"
                    min={0}
                    max={3600000}
                    value={minPublishMs}
                    onChange={(event) => setMinPublishMs(event.target.value)}
                    placeholder="Optional"
                  />
                </div>
              </div>
              <Button type="submit" disabled={!deviceId || applyState.isLoading}>
                {applyState.isLoading ? 'Sending...' : 'Apply'}
              </Button>
            </form>
            {applyState.error && <p className="mt-3 text-sm text-destructive">{applyState.error}</p>}
            {renderAck(applyState.ack)}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
