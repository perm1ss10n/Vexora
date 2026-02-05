import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { DeviceSelect } from '@/components/DeviceSelect';
import { ApiError } from '@/api/client';
import { sendCommand } from '@/api/commands';
import { useDevices, useDeviceDetail } from '@/features/devices/hooks';
import { useAuthStore } from '@/store/auth';
import { useSettingsStore } from '@/store/settings';

export function SettingsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: devices = [] } = useDevices();
  const [deviceId, setDeviceId] = useState('');
  const { data: detail } = useDeviceDetail(deviceId);
  const [intervalMs, setIntervalMs] = useState('');
  const [minPublishMs, setMinPublishMs] = useState('');
  const [applyError, setApplyError] = useState<string | null>(null);
  const [applyNotice, setApplyNotice] = useState<string | null>(null);
  const [isApplying, setIsApplying] = useState(false);
  const accessToken = useAuthStore((state) => state.accessToken);
  const refreshAccessToken = useAuthStore((state) => state.refreshAccessToken);
  const logout = useAuthStore((state) => state.logout);
  const { denseMode, glowEffects, telemetryIntervalMs, minPublishMs: uiMinPublishMs, update } = useSettingsStore();

  useEffect(() => {
    if (!deviceId && devices.length) {
      setDeviceId(devices[0].deviceId);
    }
  }, [deviceId, devices]);

  useEffect(() => {
    setApplyError(null);
    setApplyNotice(null);
  }, [deviceId]);

  const settingsKnown = detail?.settingsSource !== 'backend_default' && !!detail?.settings;

  useEffect(() => {
    if (!detail) {
      return;
    }
    if (settingsKnown) {
      setIntervalMs(String(detail.settings.telemetry.intervalMs));
      setMinPublishMs(String(detail.settings.telemetry.minPublishMs));
    } else {
      setIntervalMs('');
      setMinPublishMs('');
    }
  }, [detail, settingsKnown]);

  const handleUnauthorized = async () => {
    await logout();
    navigate('/login');
  };

  const commandErrorMessage = useMemo(() => {
    return (error: unknown) => {
      if (error instanceof ApiError) {
        return error.message || `Request failed (${error.status})`;
      }
      return (error as Error).message || 'Request failed';
    };
  }, []);

  const statusBadgeVariant = useMemo(() => {
    switch (detail?.settings?.cfgStatus) {
      case 'applied':
        return 'success';
      case 'rejected':
      case 'rolled_back':
        return 'danger';
      case 'pending':
      default:
        return 'accent';
    }
  }, [detail?.settings?.cfgStatus]);

  const handleApplyConfig = async (event: React.FormEvent) => {
    event.preventDefault();
    setApplyError(null);
    setApplyNotice(null);

    if (!deviceId || !accessToken) {
      return;
    }

    const intervalValue = Number(intervalMs);
    if (Number.isNaN(intervalValue) || intervalValue < 1000 || intervalValue > 3_600_000) {
      setApplyError('Interval must be between 1000 and 3600000 ms.');
      return;
    }
    const minPublishValue = minPublishMs.trim() === '' ? undefined : Number(minPublishMs);
    if (
      typeof minPublishValue === 'number' &&
      (Number.isNaN(minPublishValue) || minPublishValue < 0 || minPublishValue > 3_600_000)
    ) {
      setApplyError('Min publish must be between 0 and 3600000 ms.');
      return;
    }

    const telemetry: Record<string, number> = { intervalMs: intervalValue };
    if (typeof minPublishValue === 'number') {
      telemetry.minPublishMs = minPublishValue;
    }

    setIsApplying(true);
    try {
      const { ack } = await sendCommand(
        deviceId,
        'apply_cfg',
        accessToken,
        { cfg: { telemetry } },
        undefined,
        async () => refreshAccessToken(),
        handleUnauthorized
      );
      if (ack.ok) {
        setApplyNotice('Applied (pending)');
        await queryClient.invalidateQueries({ queryKey: ['device-detail', deviceId] });
      } else {
        setApplyError(ack.msg || ack.code || 'Apply failed.');
      }
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        await handleUnauthorized();
      }
      setApplyError(commandErrorMessage(error));
    } finally {
      setIsApplying(false);
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <CardTitle>Settings</CardTitle>
          <DeviceSelect value={deviceId} onChange={setDeviceId} className="w-64" />
        </CardHeader>
      </Card>

      <Card>
        <CardHeader className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <CardTitle>Telemetry cadence</CardTitle>
          <Badge variant={statusBadgeVariant as 'accent' | 'success' | 'danger'}>
            {detail?.settings?.cfgStatus ?? 'unknown'}
          </Badge>
        </CardHeader>
        <CardContent>
          <div className="mb-4 text-sm text-muted-foreground">
            Current values:{' '}
            <span className="text-foreground">
              {settingsKnown
                ? `${detail?.settings.telemetry.intervalMs ?? '—'} / ${detail?.settings.telemetry.minPublishMs ?? '—'}`
                : 'unknown'}
            </span>
          </div>
          <form className="grid gap-4 md:grid-cols-2" onSubmit={handleApplyConfig}>
            <div>
              <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">Interval (ms)</label>
              <Input
                value={intervalMs}
                onChange={(event) => setIntervalMs(event.target.value)}
                placeholder={settingsKnown ? undefined : 'unknown'}
                type="number"
                min={1000}
                max={3_600_000}
              />
              <p className="mt-2 text-xs text-muted-foreground">Частота отправки телеметрии.</p>
            </div>
            <div>
              <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">
                Min publish (ms)
              </label>
              <Input
                value={minPublishMs}
                onChange={(event) => setMinPublishMs(event.target.value)}
                placeholder={settingsKnown ? undefined : 'unknown'}
                type="number"
                min={0}
                max={3_600_000}
              />
              <p className="mt-2 text-xs text-muted-foreground">Лимитер на частоту публикации метрик.</p>
            </div>
            <div className="md:col-span-2">
              <Button type="submit" disabled={!deviceId || isApplying}>
                {isApplying ? 'Applying...' : 'Save & apply'}
              </Button>
              {applyNotice && <p className="mt-3 text-sm text-emerald-400">{applyNotice}</p>}
              {applyError && <p className="mt-3 text-sm text-destructive">{applyError}</p>}
            </div>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>UI preferences</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <label className="flex items-center justify-between gap-4 rounded-lg border border-border/70 bg-muted/20 px-4 py-3">
            <div>
              <div className="font-medium">Dense mode</div>
              <div className="text-xs text-muted-foreground">Компактное размещение таблиц и списков.</div>
            </div>
            <input
              type="checkbox"
              checked={denseMode}
              onChange={(event) => update({ denseMode: event.target.checked })}
              className="h-4 w-4 accent-cyan-400"
            />
          </label>
          <label className="flex items-center justify-between gap-4 rounded-lg border border-border/70 bg-muted/20 px-4 py-3">
            <div>
              <div className="font-medium">Glow effects</div>
              <div className="text-xs text-muted-foreground">Мягкое подсвечивание активных элементов.</div>
            </div>
            <input
              type="checkbox"
              checked={glowEffects}
              onChange={(event) => update({ glowEffects: event.target.checked })}
              className="h-4 w-4 accent-cyan-400"
            />
          </label>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Default telemetry config</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">Interval (ms)</label>
            <Input
              value={telemetryIntervalMs}
              onChange={(event) => update({ telemetryIntervalMs: Number(event.target.value) })}
            />
          </div>
          <div>
            <label className="mb-2 block text-xs uppercase tracking-wider text-muted-foreground">Min publish (ms)</label>
            <Input value={uiMinPublishMs} onChange={(event) => update({ minPublishMs: Number(event.target.value) })} />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
