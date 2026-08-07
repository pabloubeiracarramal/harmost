---
category: Agents
---
StatRow — one label/value line in an agent's detail panel.

```tsx
<div>
  <StatRow label="Hostname" value="ip-10-0-3-42.eu-west-1.compute.internal" />
  <StatRow label="Version" value="0.4.1" />
  <StatRow label="Containers" value="3" />
</div>
```

Both props are plain strings — `value` renders in mono so identifiers, versions and counts align down the column. Each row carries its own bottom border and the last one drops it (`last:border-0`), so stack them directly with no divider or gap of your own.

For a key/value line that may link somewhere, use `DetailRow` instead.
