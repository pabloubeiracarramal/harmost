---
category: Jobs
---
DetailRow — one label/value line in a job's detail panel, optionally linking out.

```tsx
<DetailRow label="Job ID" value={job.id} />
<DetailRow label="Agent" value={agent.name} link={`/agents/${job.agent_id}`} />
```

`label` and `value` are strings; passing `link` turns the value into an indigo anchor instead of plain mono text. The value truncates rather than wrapping, so long image refs and UUIDs keep the row on one line.

Self-bordering like `StatRow` — stack them with no divider of your own; the last row drops its border. `link` renders a plain `<a href>`, so use it for external or already-resolved URLs.
