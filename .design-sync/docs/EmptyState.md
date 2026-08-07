---
category: Agents
---
EmptyState — the zero-agents placeholder for the dashboard.

Takes no props. Fixed copy: "No agents yet." plus the `harmost pair <hub-url>` onboarding command in an inline code chip.

```tsx
{agents.length === 0 ? <EmptyState /> : <div className="grid gap-4">…</div>}
```

Dashed border on a translucent surface, deliberately lighter than a real Card so it reads as absence rather than content. Because the copy is hardcoded, this is the agents-list empty state specifically — it is not a generic empty state.
