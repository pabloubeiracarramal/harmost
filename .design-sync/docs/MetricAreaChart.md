---
category: Agents
---
MetricAreaChart — one metric's rolling time series with its current value as the headline.

```tsx
<MetricAreaChart dataKey="cpu" label="CPU" value={34.6} samples={samples} />
<MetricAreaChart dataKey="memory" label="Memory" value={28.4}
  sublabel="4.4 GB / 15.5 GB" samples={samples} />
```

`dataKey` selects both the series to plot and its colour, and is restricted to `'cpu' | 'memory' | 'disk'` — the three validated palette slots (blue / orange / aqua), stepped for light and dark. Do not add a fourth key.

`samples` is `MetricSample[]` (`{ t, cpu, memory, disk }`, `t` = epoch ms). The y-axis is pinned to 0–100 for all three series so the facets stay comparable; the x-axis spans `windowMs` (default 10 min) ending at the newest sample, so a partly-filled series grows in from the right.

Three of these sit side by side in `MetricsCard` — the header row holds only label and value, so capacity text goes in `sublabel`, under the plot.
