---
category: Primitives
---
ChartContainer — the recharts wrapper that binds a chart to the design tokens.

Wrap any recharts tree in it and give it a `config` mapping each dataKey to a `label` and `color`/`theme`. It injects scoped `--color-<key>` CSS vars (via `ChartStyle`) and supplies the `ResponsiveContainer`, so the chart child must be a single recharts element.

```tsx
const config = { cpu: { label: 'CPU', theme: { light: '#2a78d6', dark: '#3987e5' } } } satisfies ChartConfig;

<ChartContainer config={config} className="h-24 w-full">
  <AreaChart data={samples}>
    <Area dataKey="cpu" type="monotone" isAnimationActive={false} dot={false} />
    <ChartTooltip content={<ChartTooltipContent indicator="dot" />} />
  </AreaChart>
</ChartContainer>
```

Size it with `className` (`h-24 w-full`) — never with recharts `width`/`height`. The `--color-*` vars are scoped INSIDE the container, so sibling markup (a legend dot in a header row) cannot read them.

Also exports `ChartTooltip`, `ChartTooltipContent`, `ChartLegend`, `ChartLegendContent`, `ChartStyle`.
