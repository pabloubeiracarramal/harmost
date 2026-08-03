import { Label, PolarAngleAxis, PolarRadiusAxis, RadialBar, RadialBarChart } from 'recharts';
import { ChartContainer, type ChartConfig } from '@/shared/components/ui/chart';

const START_ANGLE = 0;
// Near-full-circle sweep (356°), leaving a small 4° gap at the bottom.
// Intentional "full ring" style gauge — change to e.g. -180 for a half-circle.
const END_ANGLE = 250;

export function RadialGauge({
  percent,
  label,
  sublabel,
}: {
  percent: number;
  label: string;
  sublabel?: string;
}) {
  const value = Math.min(Math.max(percent, 0), 100);

  // This gauge shows resource usage (CPU/Memory/Disk), where LOW is good and
  // HIGH is concerning: green at low usage, sliding through yellow to red as
  // usage climbs. Flip this (e.g. `62 + (value / 100) * 80`) if a HIGH value
  // is instead the desirable outcome, like a completion/health percentage.
  const hue = 120 - (value / 100) * 120;
  const color = `hsl(${hue}, 70%, 55%)`;

  const chartData = [{ metric: label, value, fill: color }];
  const chartConfig = {
    value: {
      label: 'Value',
    },
  } satisfies ChartConfig;

  return (
    <div className="flex flex-col items-center gap-2">
      <ChartContainer config={chartConfig} className="h-32 w-32">
        <RadialBarChart
          data={chartData}
          startAngle={START_ANGLE}
          endAngle={END_ANGLE}
          innerRadius="78%"
          outerRadius="100%"
        >
          <PolarAngleAxis type="number" domain={[0, 100]} dataKey="value" tick={false} axisLine={false} />
          <RadialBar dataKey="value" background cornerRadius={6} />
          <PolarRadiusAxis tick={false} tickLine={false} axisLine={false}>
            <Label
              content={({ viewBox }) => {
                if (viewBox && 'cx' in viewBox && 'cy' in viewBox) {
                  return (
                    <text
                      x={viewBox.cx}
                      y={viewBox.cy}
                      textAnchor="middle"
                      dominantBaseline="middle"
                      className="fill-foreground text-base font-semibold"
                    >
                      {percent != null ? `${value.toFixed(1)}%` : '–'}
                    </text>
                  );
                }
                return null;
              }}
            />
          </PolarRadiusAxis>
        </RadialBarChart>
      </ChartContainer>
      <p className="text-sm font-medium text-foreground">{label}</p>
      {sublabel && <p className="text-xs text-muted-foreground">{sublabel}</p>}
    </div>
  );
}