---
id: "0009"
title: Frontend Splits Routing (routes/) from Page Composition (pages/)
date: 2026-08-01
status: Accepted
tags: [architecture, frontend, react]
---

# ADR 0009: Frontend Splits Routing (`routes/`) from Page Composition (`pages/`)

## 1. Context and Problem Statement

[ADR 0008](./0008-frontend-feature-based-architecture.md) established the feature-based structure (`app/`, `shared/`, `features/<feature>/`). At that point `src/routes/` held TanStack Router's file routes as thin wiring (`beforeLoad` guard + `component` pointer), and each feature's `components/` held one "full page" component per route (e.g. `features/agents/components/AgentList.tsx` — fetches data, subscribes to WS, wraps `AppShell`, renders the whole route).

Two things needed fixing: the route file and its feature's "full page" component were a 1:1 pair split across two directories for no functional reason, and nothing in `features/*/components/` was actually reusable *within* its own feature except the genuinely small presentational pieces (`AgentCard`, `JobStateBadge`, `ArcGauge`, ...) — the "full page" components were composition roots masquerading as feature components.

## 2. Considered Options

* Option A: Merge each `routes/*.tsx` + its paired `features/*/components/*.tsx` into a single file under `src/pages/`, and point TanStack Router's `routesDirectory` at `src/pages`. One file owns both the route definition (`createFileRoute`, `beforeLoad`, `Route.useParams()`) and the full composition.
* Option B: Keep `routes/` and `pages/` as two separate top-level directories. `routes/` stays TanStack Router's `routesDirectory` — pure routing wiring: `createFileRoute`, `beforeLoad` guards, and (for dynamic segments) a small wrapper that calls `Route.useParams()` and passes the value down as a prop. `pages/` holds one file per route's actual composition, with **zero knowledge of TanStack Router's route-definition APIs** — no `createFileRoute`, no `Route.useParams()`. A page that needs a URL param receives it as a plain prop.

## 3. Decision

**Option B.** `vite.config.ts`'s `TanStackRouterVite({ routesDirectory: './src/routes', ... })` is unchanged from ADR 0008. `src/pages/` is a new, separate top-level directory (a sibling of `app/`, `shared/`, `features/`, `routes/`) containing pure composition components — they import feature hooks and feature components, but never `@tanstack/react-router`'s `createFileRoute`/`Route`. For the two dynamic routes (`agents/$id`, `jobs/$id`), the route file keeps a small `RouteComponent` wrapper that calls `Route.useParams()` and renders `<XxxPage id={id} />`; the page component itself just takes `{ id }: { id: string }` as a prop.

**Justification:** Merging routing and composition into one file (the initially-tried Option A) technically worked, but it meant every page component was written against TanStack Router's file-route API even though most of a page's logic has nothing to do with routing — a page calling `useAgents()` and rendering a grid doesn't need to know it's inside a `createFileRoute` callback. Keeping them separate means `pages/*` are plain, framework-router-agnostic composition components: testable without touching the router, and free to be reasoned about purely in terms of "what feature hooks does this call, what does it render." `routes/*` stays exactly what it was in ADR 0008 — thin wiring whose only job is mapping a URL to an auth guard and a page.

`AppShell` still stays in `app/`, not `shared/components/layout/`. `shared/` cannot import `features/`, and `AppShell` needs `useMe()` (a `features/auth` hook) to show the signed-in user and handle logout; making it presentational would push that call into every page that renders it. `app/` is allowed to import `features/`, so `AppShell` calling `useMe()` directly remains the one intentional exception, made once, in one file — this reasoning is unchanged from the earlier draft of this ADR.

> **Superseded 2026-08-17:** [ADR 0011](./0011-authenticated-shell-in-routes.md) replaces `AppShell` with a TanStack Router pathless layout route (`routes/_authenticated.tsx`) that now owns the `useMe()` call directly, and moves the presentational shell (`AppLayout`, `SidebarContainer`) into `shared/components/layout/`. The routing/pages split described below is otherwise unchanged.

## 4. Consequences

**Positive:**
* `pages/*` are pure — no import of `@tanstack/react-router`'s route-definition APIs (`createFileRoute`, `Route`) anywhere under `src/pages/`. Only `Link`/`useNavigate` (plain navigation, not route definition) show up where a page needs to link or redirect.
* `routes/*` stays exactly as thin as ADR 0008 left it — a guard plus a pointer to a page (or, for dynamic segments, a guard plus a 4-line params-extraction wrapper).
* `features/*/components/` holds only genuinely reusable, feature-owned presentational pieces (`AgentCard`, `EmptyState`, `ArcGauge`, `StatRow`, `JobStateBadge`, `Field`, `Input`, `DetailRow`, `LogViewer`) — not full pages. `agent-tokens` and `auth` have no `components/` directory at all, since neither had anything worth extracting.
* File paths under `src/pages/` mirror `src/routes/` (e.g. `routes/jobs/$id.tsx` ↔ `pages/jobs/$id.tsx`), so the route↔page pairing is obvious from the path alone even though `pages/` isn't parsed by the router plugin.

**Negative / Trade-offs:**
* Back to two files per non-trivial route (route wiring + page composition) instead of one — the file-count reduction from the Option-A draft of this ADR is given up in exchange for the router-knowledge boundary.
* The `$id` routes need the small `RouteComponent` wrapper again (`Route.useParams()` → prop). Minor, but it's a pattern to remember: params cross the routes/pages boundary as props, never via the page importing `Route` itself.
* `pages/` and `features/*/components/` can look similar at a glance (both hold `.tsx` UI). The distinguishing rule: `pages/` = one file per route, composes features, wraps `AppShell` (or is a full-screen unauthenticated view); `features/*/components/` = smaller pieces reused *within* rendering a page, never wraps `AppShell` itself, never imported by a route directly.

---

## 5. AI Directives (System Rules)

* **MUST:** Add new routes as files under `src/routes/` (TanStack Router's `routesDirectory`, unchanged from ADR 0008). Add the corresponding composition as a new file under `src/pages/`, at the mirrored path.
* **MUST NOT:** Import `createFileRoute`, `Route`, or call `Route.useParams()` anywhere under `src/pages/`. Pages take route params as plain props. `Link`/`useNavigate` are fine in `pages/` — those are navigation, not route definition.
* **MUST:** For a dynamic segment, keep the route file's `RouteComponent` wrapper pattern: `const { id } = Route.useParams(); return <XxxPage id={id} />;`. Don't move that extraction into the page.
* **MUST:** A page may call feature hooks (`useAgents()`, `useJobsListSocket()`, etc.) and render feature components — that's composition. It must NOT define its own `useQuery`/`useMutation`/WS subscription inline; that belongs in the owning feature's `api/`.
* **MUST NOT:** Move `AppShell` into `shared/components/layout/` without first resolving how it gets `useMe()`/session data without either (a) `shared/` importing `features/` or (b) every page duplicating that fetch. _(Superseded by [ADR 0011](./0011-authenticated-shell-in-routes.md) — resolved via a pathless layout route in `routes/`.)_
* **REFERENCE:** See `apps/front/CLAUDE.md` for the current directory layout and file-by-file conventions, and [ADR 0008](./0008-frontend-feature-based-architecture.md) for the feature/shared boundary rules this ADR doesn't change.
