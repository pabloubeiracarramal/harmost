---
category: Agents
---
AgentCard — one agent in the dashboard grid, as a navigable tile.

Takes a single `agent: Agent` and renders the whole tile itself, including the link. It is a TanStack Router `<Link to="/agents/$id">`, so it must render inside a router context.

```tsx
<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
  {agents.map((a) => <AgentCard key={a.id} agent={a} />)}
</div>
```

Reads `name`, `hostname`, `status`, and optionally `version` and `last_seen_at`. `status === 'online'` drives the emerald pill with its dot; anything else renders the neutral "Offline" pill. Pass a whole agent — there is no prop to override the label or destination.
