---
id: "0011"
title: Authenticated Shell Moves to a Pathless Layout Route
date: 2026-08-17
status: Accepted
tags: [architecture, frontend, react]
---

# ADR 0011: Authenticated Shell Moves to a Pathless Layout Route

## 1. Context and Problem Statement

[ADR 0009](./0009-frontend-pages-layer.md) established `AppShell` as a single component in `app/`, deliberately kept out of `shared/components/layout/` because it calls `useMe()` (a `features/auth` hook) to show the signed-in user and handle logout, and `shared/` cannot import `features/`. At that point every protected route file (`dashboard.tsx`, `tokens.tsx`, `agents/$id.tsx`, `jobs/*.tsx`) also carried its own identical `beforeLoad: () => { if (!isAuthenticated()) throw redirect({ to: '/login' }) }` guard, copy-pasted per file.

Redesigning the shell from a top header to a sidebar (nav + user + logout) meant splitting `AppShell` into presentational pieces (`AppLayout`, `SidebarContainer`) worth keeping in `shared/components/layout/` for reuse/testability. That resurfaced the exact problem ADR 0009 flagged: something still has to call `useMe()` and own the logout handler, and it can't live in `shared/`.

## 2. Considered Options

* Option A: Keep a single `AppShell`-equivalent file in `app/`, just restyle its internals for the sidebar. No structural change from ADR 0009.
* Option B: Extract `AppLayout`/`SidebarContainer` into `shared/components/layout/`, but keep a new dedicated wrapper file in `app/` (e.g. `app/AuthenticatedLayout.tsx`) that owns `useMe()`/logout and composes them — the literal reading of ADR 0009's "AppShell stays in `app/`" rule.
* Option C: Extract the same presentational pieces into `shared/`, and fold the `useMe()`/logout composition directly into `routes/_authenticated.tsx` — a new TanStack Router **pathless layout route** (`/_authenticated`, no path segment) that wraps `dashboard`, `tokens`, `agents/$id`, and `jobs/*` via `beforeLoad` + a nested `Outlet`. Also move the auth guard here instead of duplicating it per protected route.

## 3. Decision

**Option C.** `routes/_authenticated.tsx` defines `beforeLoad` (the one auth guard, now shared by every route nested under it) and a `component` that calls `useMe()`, builds the logout handler, and renders `<AppLayout sidebar={<SidebarContainer user={me} onLogout={handleLogout} />}><Outlet /></AppLayout>`. Every previously-protected route moved under `routes/_authenticated/` (mirroring `pages/`'s existing path structure) and lost its own `beforeLoad` — it's now a pure `createFileRoute` + `component` pointer (or, for dynamic segments, the existing `RouteComponent` + `Route.useParams()` wrapper from ADR 0009).

**Justification:** A pathless layout route's `component` *is* the shell for everything nested under it — composing `useMe()` + `AppLayout` + `SidebarContainer` there isn't business logic leaking into routing, it's what that route exists to do. Option B's dedicated `app/AuthenticatedLayout.tsx` was a file with exactly one caller (`routes/_authenticated.tsx`) and no behavior beyond what the route file could hold itself — pure indirection. The project already accepts inlining route-specific composition into `routes/*.tsx` for dynamic segments (the `RouteComponent` pattern in `agents/$id.tsx`/`jobs/$id.tsx`, which calls `Route.useParams()` inline rather than farming extraction out to `app/`); this is the same shape of exception, applied to a layout route instead of a dynamic one.

This **supersedes** the clause in ADR 0009 reserving direct `features/*` hook calls for `app/` only. That clause was written when the shell was a single flat component with no natural "route" home. Now that the auth boundary is expressed as a TanStack Router pathless layout route, the route file is the more natural home for the one `useMe()` call than a single-purpose `app/` wrapper.

## 4. Consequences

**Positive:**
* One `beforeLoad` guard (`routes/_authenticated.tsx`) instead of one copy-pasted into every protected route file.
* `AppLayout` (shell frame) and `SidebarContainer` (nav + user + logout, purely presentational — `user`/`onLogout` as props) live in `shared/components/layout/`, testable without a router or `useMe()`.
* `PageContainer` (`shared/components/layout/page-container/`) gives every authenticated page a consistent title/description/actions header, replacing each page's hand-rolled header markup.
* `app/` no longer holds a single-caller wrapper file; it's back to just `router.tsx`.

**Negative / Trade-offs:**
* `routes/_authenticated.tsx` now imports a `features/auth` hook (`useMe`) directly — the one exception to "`routes/` is pure wiring, never imports `features/`" from ADR 0008/0009. No other route file does this; unauthenticated routes (`login.tsx`, `device.tsx`, `auth/callback.tsx`) stay guard-only.
* Unauthenticated pages (`login`, `device`) get their chrome from a separate `shared/components/layout/auth-layout/AuthLayout.tsx`, not `AppLayout`. Two parallel shell components now exist by design (authenticated vs. unauthenticated chrome) rather than one `AppShell` branching internally.

---

## 5. AI Directives (System Rules)

* **MUST:** Add new protected routes under `routes/_authenticated/<path>.tsx` (mirroring the equivalent `pages/<path>.tsx`) rather than adding a per-route `beforeLoad` guard — the pathless layout route at `routes/_authenticated.tsx` already enforces `isAuthenticated()` for everything nested under it.
* **MUST:** Keep `routes/_authenticated.tsx` as the *only* file under `routes/` allowed to import a `features/*` hook directly. Every other route file stays a guard + `component` pointer, or (for dynamic segments) a guard + the `RouteComponent`/`Route.useParams()` wrapper from ADR 0009.
* **MUST NOT:** Reintroduce a single-caller `app/` wrapper file for the authenticated shell composition. If the composition grows complex enough to need splitting further, extract more presentational pieces into `shared/components/layout/`, not another `app/` indirection layer.
* **MUST NOT:** Give unauthenticated pages (`login`, `device`) the authenticated `AppLayout`/`SidebarContainer` — they use `shared/components/layout/auth-layout/AuthLayout.tsx` instead.
* **REFERENCE:** See [ADR 0009](./0009-frontend-pages-layer.md) for the `routes/` vs `pages/` split (unchanged by this ADR) and `apps/front/CLAUDE.md` for the current directory layout.
