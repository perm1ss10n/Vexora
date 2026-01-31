import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { useSettingsStore } from '@/store/settings';

export function SettingsPage() {
  const { denseMode, glowEffects, telemetryIntervalMs, minPublishMs, update } = useSettingsStore();

  return (
    <div className="space-y-6">
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
            <Input value={minPublishMs} onChange={(event) => update({ minPublishMs: Number(event.target.value) })} />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
