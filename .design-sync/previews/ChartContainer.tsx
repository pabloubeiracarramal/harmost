// recharts primitives come from the DS bundle, not from `recharts` directly:
// ChartContainer's ResponsiveContainer will not render children built by a
// second copy of recharts (the chart silently draws nothing).
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  ChartContainer,
  XAxis,
  YAxis,
} from 'front';

const jobsPerHour = [
  { hour: '06:00', succeeded: 14, failed: 1 },
  { hour: '08:00', succeeded: 22, failed: 0 },
  { hour: '10:00', succeeded: 31, failed: 3 },
  { hour: '12:00', succeeded: 27, failed: 1 },
  { hour: '14:00', succeeded: 38, failed: 2 },
  { hour: '16:00', succeeded: 24, failed: 0 },
];

const cpuSeries = Array.from({ length: 20 }, (_, i) => ({
  t: i,
  cpu: 28 + Math.sin(i / 2.2) * 9 + (i % 3) * 1.6,
}));

/** An area chart — the shape MetricAreaChart builds on. */
export function AreaSeries() {
  return (
    <div style={{ maxWidth: 560 }}>
      <ChartContainer
        config={{ cpu: { label: 'CPU', theme: { light: '#2a78d6', dark: '#3987e5' } } }}
        className="h-40 w-full"
      >
        <AreaChart data={cpuSeries} margin={{ top: 8, right: 8, bottom: 4, left: 0 }}>
          <CartesianGrid vertical={false} className="stroke-border" />
          <YAxis
            domain={[0, 100]}
            ticks={[0, 50, 100]}
            width={30}
            tickLine={false}
            axisLine={false}
            className="text-[10px] fill-muted-foreground"
          />
          <Area
            dataKey="cpu"
            type="monotone"
            stroke="var(--color-cpu)"
            fill="var(--color-cpu)"
            fillOpacity={0.18}
            strokeWidth={2}
            isAnimationActive={false}
            dot={false}
          />
        </AreaChart>
      </ChartContainer>
    </div>
  );
}

/** Two series with a legend-style config — colours come from `config`. */
export function BarSeries() {
  return (
    <div style={{ maxWidth: 560 }}>
      <ChartContainer
        config={{
          succeeded: { label: 'Succeeded', theme: { light: '#1baf7a', dark: '#199e70' } },
          failed: { label: 'Failed', theme: { light: '#eb6834', dark: '#d95926' } },
        }}
        className="h-40 w-full"
      >
        <BarChart data={jobsPerHour} margin={{ top: 8, right: 8, bottom: 4, left: 0 }}>
          <CartesianGrid vertical={false} className="stroke-border" />
          <XAxis
            dataKey="hour"
            tickLine={false}
            axisLine={false}
            className="text-[10px] fill-muted-foreground"
          />
          <Bar dataKey="succeeded" fill="var(--color-succeeded)" radius={3} isAnimationActive={false} />
          <Bar dataKey="failed" fill="var(--color-failed)" radius={3} isAnimationActive={false} />
        </BarChart>
      </ChartContainer>
    </div>
  );
}

/** Sized purely by className — never by recharts width/height. */
export function Compact() {
  return (
    <div style={{ maxWidth: 280 }}>
      <ChartContainer
        config={{ cpu: { label: 'CPU', theme: { light: '#2a78d6', dark: '#3987e5' } } }}
        className="h-20 w-full"
      >
        <AreaChart data={cpuSeries} margin={{ top: 4, right: 4, bottom: 0, left: 0 }}>
          <Area
            dataKey="cpu"
            type="monotone"
            stroke="var(--color-cpu)"
            fill="var(--color-cpu)"
            fillOpacity={0.18}
            strokeWidth={2}
            isAnimationActive={false}
            dot={false}
          />
        </AreaChart>
      </ChartContainer>
    </div>
  );
}
