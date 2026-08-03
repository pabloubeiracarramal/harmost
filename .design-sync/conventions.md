## Harmost UI — how to build with this library

Harmost is a CI/CD orchestration dashboard: agents, jobs, live logs and metrics.

### Setup: this library is dark-only

The app hardcodes `class="dark"` on `<html>` and never offers a light mode. The `.dark` block in `styles.css` supplies the real token values, and every component below is authored against a dark surface. **Put `dark` on your root element and set the base surface**, or components render dark-on-dark and read as broken:

```tsx
<div className="dark min-h-screen bg-neutral-950 text-white">
  {/* your page */}
</div>
```

Two notes on that root:

- `dark` must be on an **ancestor of everything**, including portals. Radix renders `TooltipContent` into `document.body`, so scoping `dark` to an inner wrapper leaves tooltips resolving light tokens (an invisible near-black pill on a near-black page). Putting it on `<html>` or your outermost element is the fix.
- **Do not use `HarmostPreviewShell` for layout.** It is exported only because the preview cards need it, and it carries a `margin: -24px` hack specific to the preview harness. Write the root above instead.

### Styling idiom: Tailwind v4 utilities, two vocabularies

There is no `tailwind.config.*` — theme config lives in `styles.css` under `@theme`. Style with utility classes. Two vocabularies coexist, and which one to use depends on what you are building:

**1. Semantic tokens** — used by the shadcn primitives (`Card`, `ChartContainer`, `Tooltip`). Prefer these for new surfaces:

| Purpose | Classes |
|---|---|
| Page / panel surface | `bg-background`, `bg-card`, `bg-popover` |
| Text | `text-foreground`, `text-card-foreground`, `text-muted-foreground` |
| Edges | `border-border`, `rounded-xl` (radius scale keys off `--radius`) |

**2. The raw neutral scale** — what every feature component (`AgentCard`, `LogViewer`, `StatRow`, `Input`…) actually uses. Match it when extending those:

`bg-neutral-950` (page) · `bg-neutral-900` (panel) · `bg-neutral-800` (input, chip) · `border-neutral-800` · `border-neutral-700` (dashed/empty) · `text-white` · `text-neutral-300` (label) · `text-neutral-400` (secondary) · `text-neutral-500` / `text-neutral-600` (dim)

**Accent colours carry meaning — do not repurpose them:**

`text-emerald-400` online / succeeded · `text-red-400` failed / timed out · `text-indigo-400` links and the running state · `text-sky-400` image-pull and container-setup states · `text-amber-400` stopping

Conventions worth keeping: monospace (`font-mono`) for anything machine-generated — image refs, IDs, hostnames, log output; and uppercase `tracking-widest` `text-muted-foreground` for panel eyebrow headings ("SYSTEM METRICS").

### Where the truth lives

- `styles.css` and its `@import` closure (which pulls in `_ds_bundle.css`) — the compiled Tailwind output, containing every utility and token that exists. Read it before inventing a class name; a utility that is not in there has no rule.
- `components/<group>/<Name>/<Name>.prompt.md` — per-component usage, props and gotchas. `<Name>.d.ts` is the API contract.

### Charts

`ChartContainer` takes a **recharts element** as its child, and the recharts primitives (`AreaChart`, `Area`, `BarChart`, `Bar`, `XAxis`, `YAxis`, `CartesianGrid`, `ResponsiveContainer`, …) are re-exported from this library. Import them from here — a separately installed copy of recharts will not render inside the `ResponsiveContainer` this bundle provides, and the chart silently draws nothing. Size charts with `className` (`h-24 w-full`), never with recharts `width`/`height`.

### An idiomatic composition

```tsx
<div className="dark min-h-screen bg-neutral-950 text-white">
  <div className="mx-auto max-w-5xl space-y-6 p-6">
    <div className="flex items-center justify-between">
      <h1 className="text-lg font-semibold">Jobs</h1>
      <JobStateBadge state="running" />
    </div>

    <div className="rounded-xl border border-neutral-800 bg-neutral-900 p-5">
      <StatRow label="Image" value="ghcr.io/acme/api-tests:1.4.2" />
      <StatRow label="Agent" value="build-runner-01" />
    </div>
  </div>
</div>
```

Library components own their own surface, padding and borders — `AgentCard`, `Card` and `LogViewer` each render their full container. Add only layout glue (grid, spacing, page chrome) around them.
