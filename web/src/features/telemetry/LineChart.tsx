import { TelemetryPoint } from '@/api/types';

interface LineChartProps {
  points: TelemetryPoint[];
}

export function LineChart({ points }: LineChartProps) {
  if (points.length === 0) {
    return <div className="text-sm text-muted-foreground">No data for this range.</div>;
  }

  const values = points.map((point) => point.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;
  const padding = 8;
  const chartMin = min - range * 0.05;
  const chartMax = max + range * 0.05;
  const chartRange = chartMax - chartMin || 1;

  const normalized = points.map((point, index) => {
    const x = padding + (index / Math.max(points.length - 1, 1)) * (100 - padding * 2);
    const y = padding + (1 - (point.value - chartMin) / chartRange) * (100 - padding * 2);
    return `${x},${y}`;
  });

  const formatTime = (ts: number) =>
    new Date(ts * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  const startTs = points[0]?.ts ?? 0;
  const midTs = points[Math.floor(points.length / 2)]?.ts ?? startTs;
  const endTs = points[points.length - 1]?.ts ?? startTs;

  return (
    <div className="h-64 w-full rounded-lg border border-border bg-muted/20 p-4">
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="h-full w-full">
        <defs>
          <linearGradient id="konyxLine" x1="0%" x2="100%" y1="0%" y2="0%">
            <stop offset="0%" stopColor="#22d3ee" stopOpacity="0.9" />
            <stop offset="100%" stopColor="#3b82f6" stopOpacity="0.9" />
          </linearGradient>
        </defs>
        <line x1={padding} y1={padding} x2={padding} y2={100 - padding} stroke="#1f2937" strokeWidth="0.4" />
        <line
          x1={padding}
          y1={100 - padding}
          x2={100 - padding}
          y2={100 - padding}
          stroke="#1f2937"
          strokeWidth="0.4"
        />
        <polyline
          fill="none"
          stroke="url(#konyxLine)"
          strokeWidth="2.2"
          strokeLinejoin="round"
          strokeLinecap="round"
          points={normalized.join(' ')}
        />
        <text x={padding} y={100 - 1} fontSize="4" textAnchor="start" fill="#6b7280">
          {formatTime(startTs)}
        </text>
        <text x={50} y={100 - 1} fontSize="4" textAnchor="middle" fill="#6b7280">
          {formatTime(midTs)}
        </text>
        <text x={100 - padding} y={100 - 1} fontSize="4" textAnchor="end" fill="#6b7280">
          {formatTime(endTs)}
        </text>
        <text x={padding - 2} y={padding + 2} fontSize="4" textAnchor="end" fill="#6b7280">
          {chartMax.toFixed(1)}
        </text>
        <text x={padding - 2} y={100 - padding} fontSize="4" textAnchor="end" fill="#6b7280">
          {chartMin.toFixed(1)}
        </text>
      </svg>
    </div>
  );
}
