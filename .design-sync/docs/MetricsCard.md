---
category: Agents
---
MetricsCard — the full system-metrics panel for one agent (CPU, memory, disk).

```tsx
<MetricsCard agent={agent} />
```

Takes only `agent: Agent` and composes everything itself: a `Card`, three `MetricAreaChart`s in a responsive grid, a tooltip on the "Updated …" timestamp, and a footer counting running containers (hidden when zero).

Percentages are derived internally from `*_used_bytes` / `*_total_bytes`, and bytes are humanised (`4.4 GB / 15.5 GB`). The series comes from `useMetricsHistory(agent)`, which buffers samples **per mount** as `last_seen_at` moves — the hub keeps no metrics history, so a freshly mounted card shows the "Collecting samples" hint until a second heartbeat lands.

Drop it straight into an agent detail page; there is nothing to wire up beyond passing the agent.
