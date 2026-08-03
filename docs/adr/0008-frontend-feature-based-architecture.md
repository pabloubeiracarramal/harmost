---
id: "0008"
title: Frontend Restructured to Feature-Based Architecture with a Shared Kernel
date: 2026-08-01
status: Accepted
tags: [architecture, frontend, react]
---

# ADR 0008: Frontend Restructured to Feature-Based Architecture with a Shared Kernel

> **Note:** [ADR 0009](./0009-frontend-pages-layer.md) refines the page-composition detail of this decision — the "full page" components originally described as living in `features/*/components/` below were extracted into a new top-level `src/pages/` directory (kept separate from `src/routes/`, which still holds TanStack Router's file-route wiring). `features/*/components/` now holds only smaller, genuinely reusable presentational pieces. The feature/shared boundary rules below are unchanged.

## 1. Context and Problem Statement

The front app (`apps/front`) started as flat, technical-type folders: `src/routes/` (TanStack Router file routes, each embedding its own page markup), `src/components/`, `src/hooks/`, `src/lib/`. As the app grew past the first few routes, two concrete problems showed up:

- The same TanStack Query call (`queryKey: ['agents']`, `api.get('/api/v1/agents')`) was hand-duplicated verbatim in four different route files (dashboard, jobs list, new-job form, tokens list) — no single place owned "how do you fetch agents."
- `useHubEvents` (the WebSocket hook) opened a brand-new `WebSocket` per call site. It happened to work because the router only ever mounts one page at a time, but nothing actually enforced a single connection — it wasn't a real single-connection guarantee, just a coincidence of routing.

An intermediate step split each route file into a thin route (`beforeLoad` + `component`) plus a page component under `src/pages/`, which fixed route-file bloat but didn't address either problem above, and page components still called `useQuery`/`useHubEvents` directly with hand-typed keys.

## 2. Considered Options

* Option A: Keep the technical-type layout (`pages/`, `components/`, `hooks/`, `lib/`) and just extract query hooks into `src/hooks/`, named by domain entity (`useAgents.ts`, `useJobs.ts`).
* Option B: Adopt a feature-based structure with a shared kernel — `src/shared/` (http client, WS client, generic UI/lib) never imports from features; `src/features/<feature>/` owns its `api/{queries,mutations,keys,socket}.ts` and `components/`, exposed only through `index.ts`; `src/app/` owns the router, top-level providers, and app shell.

## 3. Decision

**Option B.** The frontend is restructured into `app/`, `shared/`, and `features/{agents,jobs,agent-tokens,auth}/`.

**Justification:** Option A would have fixed the query-key duplication but not the deeper issue — there's no place that owns "this is everything the agents domain knows about," so cross-cutting duplication tends to reappear as the app grows. Option B fixes it at the structural level: a query key factory is mandatory per feature (`agentKeys`, `jobKeys`, `agentTokenKeys`, `authKeys`), so a key is written once and reused for both fetching and cache invalidation. It also gave a natural place to fix the WebSocket connection for real: `shared/ws/wsClient.ts` is a true singleton (reference-counted subscribe/unsubscribe, connects lazily, disconnects when the last subscriber leaves), and each feature's `api/socket.ts` only translates the events it cares about into cache writes via `shared/ws/useWsSubscribe.ts`. Live, append-only streams (job logs) are explicitly kept out of the query cache per rule 3 of the adopted convention — that was already how `JobDetail`'s log viewer worked, which confirmed the convention fit the app rather than fighting it.

Cross-feature data needs (e.g. `jobs` and `agent-tokens` both need `useAgents()` to resolve a name) are resolved by importing the other feature's public `index.ts` — never its internals. `app/AppShell.tsx` is the one place allowed to reach into a feature (`features/auth`'s `useMe()`) from outside `features/`, since `shared/` cannot import from `features/` but `app/` can.

## 4. Consequences

**Positive:**
* Query keys are defined once per feature and reused for fetching and invalidation — the four-way `['agents']` duplication is gone.
* The hub WebSocket connection is now an actual singleton with reference-counted lifecycle, not an accident of routing.
* Each feature's public surface is exactly its `index.ts` — `git grep` for cross-feature reach-ins is a straightforward way to audit boundary violations even without lint tooling in place yet.
* `src/app/` (router, `AppShell`, top-level providers) replaced ~900 lines of unused Nx scaffold (`nx-welcome.tsx` and friends) that had been sitting there unreferenced since the app was first generated.

**Negative / Trade-offs:**
* More files for the same amount of logic — e.g. `agent-tokens` now has 4 small files (`types.ts`, `keys.ts`, `queries.ts`, `mutations.ts`) for what was one file's worth of code before. Justified at this size mainly by consistency across features, not by per-file necessity yet.
* Rule 8 of the adopted convention (lint-enforced boundaries via `eslint-plugin-boundaries` or `import/no-restricted-paths`) is **not** implemented — there is no ESLint configuration anywhere in this repo yet. Boundaries (`shared/` never importing `features/`, features only importing each other via `index.ts`) are currently a convention enforced by review, not tooling.
* `shared/ws/wsClient.ts` intentionally types WS job-status payloads loosely (`state: string`, not the `JobState` union) to avoid `shared/` depending on `features/jobs`'s domain type. `features/jobs/api/socket.ts` casts (`as JobState`) at the one point it writes that value into the `Job` cache — a deliberate, narrow boundary cast, not a general pattern to copy elsewhere.

---

## 5. AI Directives (System Rules)

* **MUST:** Put new TanStack Query hooks in the owning feature's `api/queries.ts` or `api/mutations.ts`, never inline in a component. Add/extend that feature's `api/keys.ts` factory rather than hand-writing a `queryKey` array.
* **MUST:** Route all WebSocket consumption through `shared/ws/useWsSubscribe.ts` (backed by the `shared/ws/wsClient.ts` singleton). Never call `new WebSocket(...)` outside `shared/ws/wsClient.ts`.
* **MUST:** Keep append-only live streams (e.g. job logs) out of the TanStack Query cache — buffer them in local component state (or a feature-local store), seeded from a REST snapshot when one exists, per `features/jobs/components/JobDetail.tsx`.
* **MUST:** Cross-feature imports go through the target feature's `index.ts` only (e.g. `import { useAgents } from '@/features/agents'`). Never import a path under another feature's `api/` or `components/` directly.
* **MUST NOT:** Import from `features/**` inside anything under `shared/`. If a feature's data is needed at the composition-root level (e.g. `AppShell` needing `useMe()`), that composition belongs in `src/app/`, not `src/shared/`.
* **MUST NOT:** Promote a hook, type, or component to `shared/` on first use. Per rule 7 of the adopted convention, it stays in the feature that uses it until a second, different feature needs it too.
* **REFERENCE:** See `apps/front/CLAUDE.md` for the full directory layout and file-by-file conventions this ADR establishes.
