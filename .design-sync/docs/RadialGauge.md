---
category: Agents
---
RadialGauge — a single resource-usage percentage as a near-full ring.

```tsx
<RadialGauge percent={34.6} label="CPU" sublabel="8 cores" />
```

`percent` is clamped to 0–100 and printed at one decimal in the centre; `label` sits under the ring and `sublabel` is optional secondary text.

The fill colour is derived from the value, not passed in: hue slides 120° (green) → 0° (red) as usage climbs, so **low is good**. For a metric where HIGH is the desirable outcome (health, completion), this component is the wrong choice — its colour scale would invert the meaning.

Fixed 128×128 (`h-32 w-32`) with a 356° sweep. For a time series rather than a point-in-time value, use `MetricAreaChart`.
