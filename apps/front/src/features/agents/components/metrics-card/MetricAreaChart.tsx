import { useId, type CSSProperties } from 'react';
import { Area, AreaChart, XAxis, YAxis } from 'recharts';
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/shared/components/ui/chart';
import { METRICS_WINDOW_MS, type MetricSample } from '../../hooks/useMetricsHistory';

/**
 * Categorical slots 1–3 of the validated data-viz palette (blue / orange /
 * aqua), stepped per mode. These three clear the all-pairs CVD and
 * normal-vision floors on both surfaces; don't extend the set by inventing a
 * fourth hue.
 */
export const METRIC_COLORS = {
  cpu: { light: '#2a78d6', dark: '#3987e5' },
  memory: { light: '#eb6834', dark: '#d95926' },
  disk: { light: '#1baf7a', dark: '#199e70' },
} as const;

const timeFormatter = new Intl.DateTimeFormat(undefined, {
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
});

interface MetricAreaChartProps {
  /** Which series of the sample to plot. */
  dataKey: keyof typeof METRIC_COLORS;
  label: string;
  /** Current value, in percent — shown as the headline figure. */
  value: number;
  /** e.g. "4.4 GB / 15.5 GB". */
  sublabel?: string;
  samples: MetricSample[];
  /** Width of the rolling window the x-axis spans. */
  windowMs?: number;
}

/**
 * One metric as a stat tile: label, headline percentage, capacity sublabel and
 * a plot of the rolling window.
 *
 * The three tiles are small multiples on one shared 0–100 scale — each carries
 * a single series (so no legend) and a flat 2% line has to *read* as flat
 * rather than being rescaled to fill its box.
 *
 * Which is exactly why the plot gets a recessed bed and span labels rather
 * than being left as a bare sparkline: an idle agent's disk line is dead flat
 * for hours, and with nothing to sit on it reads as a stray rule instead of a
 * chart. The bed draws the plot's bounds; the span labels are the only thing
 * on the card that says the x-axis is time. Nothing is drawn *inside* the bed
 * — no gridlines, no tick labels: three 0/50/100 gutters side by side re-crowd
 * the card and indent each plot away from its own headline, and the bed
 * already bounds the 0–100 range on its own. The headline carries the value.
 */
export function MetricAreaChart({
  dataKey,
  label,
  value,
  sublabel,
  samples,
  windowMs = METRICS_WINDOW_MS,
}: MetricAreaChartProps) {
  // Gradient ids are document-global; two charts on one page would otherwise
  // share (and fight over) the same <linearGradient>.
  const gradientId = `metric-fill-${dataKey}-${useId().replace(/:/g, '')}`;

  // `theme` here only feeds the tooltip's indicator swatch — ChartStyle scopes
  // its `--color-*` vars to inside ChartContainer, so the header dot below
  // can't use them and reads --metric off this component's own wrapper.
  const chartConfig = {
    [dataKey]: { label, theme: METRIC_COLORS[dataKey] },
  } satisfies ChartConfig;

  const color = 'var(--metric)';
  const latest = samples.at(-1)?.t;

  return (
    <div
      className="rounded-lg border bg-muted/40 p-3 [--metric:var(--metric-light)] dark:[--metric:var(--metric-dark)]"
      style={
        {
          '--metric-light': METRIC_COLORS[dataKey].light,
          '--metric-dark': METRIC_COLORS[dataKey].dark,
        } as CSSProperties
      }
    >
      <div className="flex items-center gap-2">
        <span className="h-1.5 w-1.5 shrink-0 rounded-full" style={{ backgroundColor: color }} />
        <span className="truncate text-xs font-medium uppercase tracking-wider text-muted-foreground">
          {label}
        </span>
      </div>

      {/* Proportional figures, not tabular: a standalone display-size number,
          not a column that has to align vertically. */}
      <div className="mt-2 flex items-baseline gap-0.5">
        <span className="text-2xl font-semibold leading-none text-foreground">
          {value.toFixed(1)}
        </span>
        <span className="text-sm font-medium text-muted-foreground">%</span>
      </div>

      {/* Fixed height: CPU has no capacity to report, and without the reserved
          line its sparkline would sit a row higher than the other two. */}
      <p className="mt-1 h-4 truncate text-xs text-muted-foreground">{sublabel}</p>

      <ChartContainer config={chartConfig} className="mt-2 h-14 w-full rounded-md bg-background/60">
        {/* Top margin clears the hover dot and its ring at 100%; the bottom
            one keeps a near-zero series (an idle agent's disk) off the bed's
            own bottom edge, where it would read as part of the frame. */}
        <AreaChart data={samples} margin={{ top: 6, right: 1, bottom: 3, left: 1 }}>
          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity={0.18} />
              <stop offset="100%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          {/* Anchor the axis to the full window ending at the newest sample, so
              a partly-filled series grows in from the right instead of always
              stretching edge to edge. */}
          <XAxis
            dataKey="t"
            type="number"
            domain={latest ? [latest - windowMs, latest] : ['dataMin', 'dataMax']}
            hide
          />
          {/* All three series are % of capacity, so they share one fixed 0–100
              scale — the tiles stay comparable and a 3% line reads as low. */}
          <YAxis domain={[0, 100]} hide />
          {/* One line, not the default stack. ChartTooltipContent's own
              `min-w-[8rem]` plus a label row above a value row covered most of
              a 205×56 plot, so the time and the value are folded into a single
              formatter row and the min-width is dropped. The tile header
              already says which series this is, so the label row and the
              indicator swatch are both redundant here. */}
          <ChartTooltip
            cursor={{ strokeWidth: 1 }}
            content={
              <ChartTooltipContent
                hideLabel
                hideIndicator
                className="min-w-0 gap-0 rounded-md px-2 py-1 shadow-md"
                formatter={(v, _name, item) => (
                  <span className="whitespace-nowrap text-[11px] text-muted-foreground">
                    {timeFormatter.format(new Date(item?.payload?.t ?? Date.now()))}
                    {' · '}
                    <span className="font-medium text-foreground">
                      {Number(v).toFixed(1)}%
                    </span>
                  </span>
                )}
              />
            }
          />
          <Area
            dataKey={dataKey}
            type="monotone"
            stroke={color}
            strokeWidth={2}
            strokeLinecap="round"
            strokeLinejoin="round"
            fill={`url(#${gradientId})`}
            isAnimationActive={false}
            dot={false}
            activeDot={{ r: 4, strokeWidth: 2, className: 'stroke-card' }}
          />
        </AreaChart>
      </ChartContainer>

      {/* Derived from windowMs, not hardcoded — the x domain is anchored to the
          same value, so a caller narrowing the window must not be left with a
          span label that lies about it. */}
      <div className="mt-1 flex justify-between text-[10px] text-muted-foreground">
        <span>{Math.round(windowMs / 60_000)}m</span>
        <span>now</span>
      </div>
    </div>
  );
}
