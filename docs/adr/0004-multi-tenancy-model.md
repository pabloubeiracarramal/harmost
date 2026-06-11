---
id: "0004"
title: Multi-Tenancy Model — Org as Tenant with Personal Org on Signup
date: 2026-06-11
status: Accepted
tags: [architecture, database, auth, multi-tenancy]
---

# ADR 0004: Multi-Tenancy Model — Org as Tenant with Personal Org on Signup

## 1. Context and Problem Statement

Harmost is a SaaS product intended to support teams and organizations, not just individual users. All resources (agents, jobs, logs) must be scoped to a tenant boundary so that one user's data is never visible to another unrelated user.

The choice of tenant unit has long-term consequences for the PostgreSQL schema, agent pairing, API authorization middleware, and the frontend's routing model. Getting this wrong means an expensive migration later.

## 2. Considered Options

* **Option A: User = tenant** — Each user is their own isolated tenant. Simple to build, but migrating to an org model later requires backfilling an `orgs` table and re-scoping all foreign keys.
* **Option B: Org = tenant, personal org auto-created on signup** — Users belong to orgs. On signup, a personal org is created automatically (1:1). All resources are scoped to `org_id`. Adding multi-member orgs later only requires building the invite/membership UI — the schema is already correct.
* **Option C: Full org model with invite flow from day one** — Correct long-term shape but highest initial build cost; invite/membership UI is non-trivial and not needed for the MVP.

## 3. Decision

**Org is the tenant unit.** On user signup, a personal org is auto-created and the user is set as its owner. All resources (agents, jobs, logs) are scoped to `org_id`, never to `user_id` directly.

**Justification:** Option B gives the correct long-term data model at nearly the same cost as Option A. The extra work is one additional table and one join in auth middleware. The payoff is zero migration cost when multi-member orgs are added. This is the model used by GitHub, Linear, and Stripe for the same reasons.

## 4. Consequences

**Positive:**
* Schema is correct from day one — no backfill migration when orgs become multi-member.
* Agent pairing is scoped to an org token, so new org members automatically see all existing agents without re-pairing.
* Authorization middleware has a single, consistent scope check (`org_id`) across all resources.
* Personal and team orgs are handled identically by the backend.

**Negative / Trade-offs:**
* Every query that touches a resource must join through `org_id` — slightly more complex than user-scoped queries.
* Signup flow must create two rows (user + org + membership) in a transaction.
* Frontend routing must be org-aware (e.g., `/orgs/:slug/agents`) from the start.

### Canonical Schema Shape

```sql
users        (id, email, ...)
orgs         (id, slug, name, personal BOOL, ...)
org_members  (org_id, user_id, role)   -- role: owner | member | viewer
agents       (id, org_id, name, ...)
jobs         (id, org_id, agent_id, ...)
job_logs     (id, job_id, ...)
```

---

## 5. AI Directives (System Rules)

* **MUST:** All resources (agents, jobs, job_logs, and any future domain entities) MUST have an `org_id` foreign key. Never scope a resource directly to `user_id`.
* **MUST:** On user signup, create a personal org and an `org_members` row (role: `owner`) in the same database transaction as the user row.
* **MUST:** Agent pairing tokens MUST be scoped to an org, not to a user.
* **MUST NOT:** Never add a resource table that references `user_id` as its top-level ownership boundary. Route ownership through `org_id → org_members → user_id`.
* **REFERENCE:** See `docs/architecture.md` for the overall data flow and `docs/adr/0002-inter-service-communication.md` for agent pairing protocol details.
