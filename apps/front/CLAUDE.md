# Front App — AI Context

## What this app does
Web UI for Harmost: GitHub OAuth login, agent dashboard with live status, agent detail with live metric area charts, device-flow approval for agent pairing, and jobs (dispatch form, list with live state badges, detail with live log viewer). Talks to the hub via REST (JWT bearer) for commands and a WebSocket (`/ws?token=<jwt>`) for live events; the Vite dev server proxies `/auth`, `/api`, and `/ws` to the hub on :8080.

## Architecture: feature-based with a shared kernel

See [ADR 0008](../../docs/adr/0008-frontend-feature-based-architecture.md) and [ADR 0009](../../docs/adr/0009-frontend-pages-layer.md) for the full rationale. Directory layout:

```
src/
  app/                    # router, top-level providers, app shell
    router.tsx            # createRouter() + Register
    AppShell.tsx           # authenticated chrome (nav, /me identity, sign-out)
  shared/                 # never imports from features/, pages/, or routes/
    api/
      httpClient.ts        # the one fetch wrapper (JWT bearer, 401 handling)
      auth.ts               # token storage (getToken/setToken/clearToken/isAuthenticated)
      queryClient.ts        # the one QueryClient instance
    ws/
      wsClient.ts           # singleton WS connection (reference-counted subscribe)
      useWsSubscribe.ts      # generic hook wrapping wsClient.subscribe
    types/api-error.ts     # ApiError
    lib/utils.ts           # cn() helper
    components/            # generic UI primitives (ui/, layout/) — none yet
  features/
    agents/                # agent domain: list/detail hooks, socket, presentational pieces
    jobs/                  # job domain: list/dispatch/detail hooks, socket, presentational pieces
    agent-tokens/          # agent token domain: list/revoke hooks
    auth/                  # session identity (useMe)
  routes/                 # TanStack Router file routes — routing wiring ONLY
  pages/                  # one file per route — composition ONLY, no router-definition APIs
```

`routes/` and `pages/` are deliberately separate (see ADR 0009):
- **`routes/`** is TanStack Router's `routesDirectory` (set in `vite.config.ts`). Each file: `createFileRoute(path)`, a `beforeLoad` auth guard, and a `component` pointing at the matching file in `pages/`. For dynamic segments (`agents/$id.tsx`, `jobs/$id.tsx`), the route keeps a small `RouteComponent` wrapper that calls `Route.useParams()` and renders `<XxxPage id={id} />` — params cross into `pages/` as a plain prop.
- **`pages/`** mirrors `routes/`'s paths (e.g. `routes/jobs/$id.tsx` ↔ `pages/jobs/$id.tsx`) but contains **zero knowledge of TanStack Router's route-definition APIs** — no `createFileRoute`, no `Route`, no `Route.useParams()`. A page:
  - calls feature hooks directly (`useAgents()`, `useJobsListSocket()`, etc.) — this is *composing* a feature, not owning business logic
  - renders feature-owned presentational components (`AgentCard`, `JobStateBadge`, ...) plus its own layout markup
  - does **not** define its own `useQuery`/`useMutation`/WS subscription — if a page needs one that doesn't exist yet, it goes in the owning feature's `api/`, not inline in the page
  - may use `Link`/`useNavigate` freely (plain navigation, not route definition)
- Trivial routes with no real composition behind them (`routes/index.tsx` — home redirect, `routes/__root.tsx` — router root layout, `routes/auth/callback.tsx` — a spinner + redirect effect) have no `pages/` counterpart; they're just the guard/effect, entirely in `routes/`.

Each `features/<feature>/` follows:
- `api/types.ts` — domain types (e.g. `Agent`, `Job`), **re-exported from the generated `@/shared/api/schema`**, never hand-written. Add a field by editing `libs/harmost-api/openapi.yaml` and running `nx run harmost-api:generate` (ADR 0010). Runtime values keyed on those types (`TERMINAL_JOB_STATES`, `isTerminal`) still live here.
- `api/keys.ts` — query key factory (`all -> lists -> list() -> details -> detail(id)`); **never** hand-write a `queryKey` array elsewhere
- `api/queries.ts` — `useQuery` hooks (GET)
- `api/mutations.ts` — `useMutation` hooks (POST/PATCH/DELETE)
- `api/socket.ts` — WS subscriptions (via `shared/ws/useWsSubscribe`) that write into the query cache; live append-only streams (job logs) stay **out** of the cache — buffered in local component state instead, seeded from a REST snapshot (see `src/pages/jobs/$id.tsx`)
- `components/` — feature-owned *presentational* UI, composed by pages (not full pages themselves — see ADR 0009)
- `hooks/` — feature-owned non-`api/` hooks: derived/buffered state that isn't a query, mutation, or raw subscription (e.g. `agents/hooks/useMetricsHistory.ts`, which accumulates heartbeat snapshots into a rolling 10-minute time series because the hub stores no metrics history)
- `index.ts` — the **only** surface other features/pages may import from this feature

**Boundaries** (not lint-enforced yet — no ESLint config exists in this repo; enforced by review):
- `shared/` never imports from `features/`, `pages/`, or `routes/`.
- Features import `shared/` freely, and other features **only** via that feature's `index.ts` (e.g. `import { useAgents } from '@/features/agents'` — `jobs` and `agent-tokens` both do this to resolve agent names).
- `pages/*` compose feature hooks/components; they don't define their own query/mutation/socket logic, and they don't import anything from `@tanstack/react-router`'s route-definition surface (`createFileRoute`, `Route`). If a page accumulates real data-fetching logic, that logic belongs in the owning feature's `api/`.
- Don't promote something to `shared/` on first use — it stays in the feature until a second, different feature needs it too.
- `AppShell` (in `app/`, not `shared/`) is the one place outside `pages/` allowed to call a feature hook (`useMe()` from `features/auth`) directly — `app/`, unlike `shared/`, is allowed to import from `features/`. This was a deliberate deviation from putting the shell under `shared/components/layout/`: doing so would require `AppShell` to take `user`/`onLogout` as props, which would push a `useMe()` + logout call into every single authenticated page instead of one place. See ADR 0009.

## Routes
- `/` — home; `/login` — GitHub OAuth button; `/auth/callback` — stores JWT, redirects
- `/dashboard` — agent list (REST) + live status (WebSocket)
- `/agents/$id` — agent detail with live metric area charts
- `/jobs` — jobs list with live state badges; `/jobs/new` — dispatch form; `/jobs/$id` — job detail + live log viewer
- `/device?code=XXXX` — device-flow approval page for agent pairing

## Runtime
- Framework: **React 19 + TypeScript 5.7**
- Bundler: **Vite** (dev server on `localhost:4200`)
- Test runner: **Vitest** (jsdom environment)
- Styling: **Tailwind CSS v4** via `@tailwindcss/vite` — no `tailwind.config.*`, config lives in `src/styles.css`
- Components: **shadcn/ui** (new-york style, zinc base, CSS variables) — add components with `npx shadcn@latest add <component>` from `apps/front/`; `components.json` aliases point at `shared/` (see below)
- Routing: **TanStack Router** (file-based) — route files live in `src/routes/`; matching page composition lives in `src/pages/` (see Architecture above). `src/routeTree.gen.ts` is auto-generated by the Vite plugin, do not edit manually
- Data fetching: **TanStack Query** — the singleton `QueryClient` lives in `src/shared/api/queryClient.ts`; `src/app/router.tsx` passes it as router context

## Nx Targets

| Target | Command |
|--------|---------|
| `nx run front:dev` (or `serve`) | Vite dev server on port 4200 |
| `nx run front:build` | Production build → `dist/apps/front` |
| `nx run front:test` | Vitest (watch=false) |
| `nx run front:typecheck` | `tsc -p tsconfig.app.json --noEmit` — the only gate that catches generated-type drift |

**MUST** use `nx run front:<target>` — never call `vite` or `tsc` directly.

## Key File Locations
- `src/routes/__root.tsx` — root layout, defines router context type
- `src/routes/index.tsx` — home route (`/`)
- `src/routeTree.gen.ts` — auto-generated route tree (do not edit)
- `src/main.tsx` — Vite entry point (required by `index.html`); thin — just mounts `<RouterProvider>`/`<QueryClientProvider>` using `app/router.tsx` and `shared/api/queryClient.ts`
- `src/app/router.tsx`, `src/app/AppShell.tsx` — see Architecture above
- `src/shared/` — see Architecture above
- `src/features/` — see Architecture above
- `src/pages/` — see Architecture above
- `components.json` — shadcn configuration (aliases point at `src/shared/`)
- `vite.config.ts` — `TanStackRouterVite({ routesDirectory: './src/routes', ... })`

## Path Alias
`@/` maps to `src/` — use it for all internal imports (e.g. `@/shared/lib/utils`, `@/features/agents`, `@/pages/dashboard`).

## Conventions
- Source root: `apps/front/src/`
- Tests: `src/**/*.spec.{ts,tsx}`
- Prefer `@testing-library/react` for component tests.
- Use `cn()` from `@/shared/lib/utils` for conditional classNames.
- Add new routes as a pair: `src/routes/<path>.tsx` (guard + `component` pointer, or a `RouteComponent` wrapper + `Route.useParams()` for dynamic segments) and `src/pages/<path>.tsx` (the actual composition, taking any params as props). The Vite plugin picks up the `routes/` file automatically on next serve/build.
- A page never imports `createFileRoute`/`Route` from `@tanstack/react-router` — only `Link`/`useNavigate` if it needs to navigate. If a page seems to need `Route.useParams()`, that's a sign the param should be threaded from the route file as a prop instead.
- New query/mutation hooks go in the owning feature's `api/` folder, never inline in a page. Check `api/keys.ts` for an existing key before writing a new one.
- New WS handling goes through `shared/ws/wsClient.ts` (via `useWsSubscribe`), never a raw `new WebSocket(...)`.

## Known Gotchas
- `routeTree.gen.ts` is regenerated on every Vite start/build. If typecheck runs before a build, the file is still valid as a placeholder — do not delete it.
- Tailwind v4 has no `tailwind.config.ts`. All theme customization goes in `src/styles.css` under `@theme`.
- `src/shared/api/schema.d.ts` is generated from `libs/harmost-api/openapi.yaml` by `nx run harmost-api:generate` (ADR 0010). Do not edit it, and do not run the generator from here — it is the `harmost-api` project's target.
- `features/agents/api/socket.ts` narrows with `e.type.startsWith('agent.')`, which does **not** narrow the union in TypeScript. It compiles only because `agent_id` exists on all three event members; a new event type without `agent_id` will break it.

## Connections to Other Apps
- Backend: **hub** — see [architecture overview](../../docs/architecture.md) for API details.
