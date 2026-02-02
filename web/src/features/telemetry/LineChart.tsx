import { TelemetryPoint } from '@/api/types';

interface LineChartProps {
  points: TelemetryPoint[];
}

export function LineChart({ points }: LineChartProps) {
  if (points.length === 0) {
    return <div className="text-sm text-muted-foreground">No telemetry data.</div>;
  }

  const values = points.map((point) => point.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;

  const normalized = points.map((point, index) => {
    const x = (index / (points.length - 1)) * 100;
    const y = 100 - ((point.value - min) / range) * 100;
    return `${x},${y}`;
  });

  return (
    <div className="h-64 w-full rounded-lg border border-border bg-muted/20 p-4">
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="h-full w-full">
        <defs>
          <linearGradient id="konyxLine" x1="0%" x2="100%" y1="0%" y2="0%">
            <stop offset="0%" stopColor="#22d3ee" stopOpacity="0.9" />
            <stop offset="100%" stopColor="#3b82f6" stopOpacity="0.9" />
          </linearGradient>
        </defs>
        <polyline
          fill="none"
          stroke="url(#konyxLine)"
          strokeWidth="2.2"
          strokeLinejoin="round"
          strokeLinecap="round"
          points={normalized.join(' ')}
        />
      </svg>
    </div>
  );
}
