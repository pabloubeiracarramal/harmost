---
id: "0005"
title: Hub Internal Structure and Tooling
date: 2026-06-11
status: Accepted
tags: [architecture, golang, hub, database, tooling]
---

# ADR 0005: Hub Internal Structure and Tooling

## 1. Context and Problem Statement

The hub is a Go backend with two distinct transports (gRPC for agents, HTTP/WebSocket for the frontend) and a PostgreSQL database. Before writing any application code, we need to decide on:

1. The internal package structure — how transport, business logic, and persistence concerns are separated.
2. The tooling stack — HTTP router, ORM/query layer, migrations, proto generation, and dev workflow.

Without these decisions recorded, contributors and AI agents will make inconsistent choices that are expensive to reconcile later.

## 2. Considered Options

**Package structure:**
* **Option A: Layered (horizontal slices)** — `transport/` → `service/` → `repository/` → `domain/`. Dependencies flow strictly inward. Transport layer is subdivided by protocol (`httpapi/`, `grpcapi/`).
* **Option B: Vertical slices (domain packages)** — one package per domain (agent, job, org), each owning its own handler/service/repo. Breaks down when one transport (the gRPC stream) crosses multiple domain boundaries simultaneously.
* **Option C: Ports & adapters (hexagonal)** — explicit interface ports for every dependency. Maximum testability but high boilerplate in Go; overkill for this stage.

**ORM / query layer:**
* **GORM** — full ORM, no SQL required, large ecosystem, fast to start. Known footguns: silent N+1 queries, non-idiomatic error handling, `AutoMigrate` is dangerous in production.
* **Ent** — schema-as-code ORM, strong type safety, `CreateBulk` for batch inserts. Steep learning curve, large generated file tree.
* **sqlc** — generates Go from hand-written SQL. No SQL authoring overhead removed; rejected because the goal is to avoid writing SQL.

**HTTP router:**
* **chi** — lightweight, `net/http`-compatible, idiomatic middleware chain.
* **Gin / Echo** — heavier frameworks with their own context types; creates friction with WebSocket libraries.

**Migrations:**
* **goose** — SQL-file based migration runner, no runtime magic, CI-friendly.
* **GORM AutoMigrate** — convenient in dev, unsafe in production; not used as the primary migration mechanism.

## 3. Decision

- **Package structure:** Option A — layered architecture.
- **ORM:** GORM
- **Migrations:** goose (AutoMigrate disabled in production)
- **HTTP router:** chi
- **Proto generation:** buf
- **Dev live reload:** air
- **Test assertions:** testify

**Justification:** Layered architecture is the natural fit for a service with two protocols — `transport/grpcapi/` and `transport/httpapi/` make the protocol boundary explicit while sharing the same service and repository layers. GORM is chosen over sqlc because authoring raw SQL is undesirable; its footguns are manageable with discipline (goose for migrations, `CreateInBatches` for bulk log writes). Chi is the minimal, idiomatic HTTP router that imposes no friction on WebSocket or OAuth library integration.

## 4. Consequences

**Positive:**
* Clear separation: transport code never contains business logic; repository code never contains HTTP concepts.
* `transport/grpcapi/` and `transport/httpapi/` can evolve independently.
* GORM handles standard CRUD with no SQL authoring; goose keeps migrations safe and version-controlled.
* Air enables fast dev iteration without manual restarts.

**Negative / Trade-offs:**
* GORM's `CreateInBatches` must be used explicitly for log chunk writes — a plain `Create` loop is a known performance footgun.
* `AutoMigrate` must never run in production; all schema changes go through goose migration files.
* Layered architecture adds indirection — a simple feature touches handler, service, and repository files.

### Canonical Package Layout

```
apps/hub/
  cmd/hub/
    main.go               # entry point, wires everything together
  internal/
    domain/               # shared types: Agent, Job, Org, User, JobLog
    transport/
      httpapi/            # chi router, WebSocket upgrade, OAuth handlers, webhook handlers
      grpcapi/            # gRPC server, bidirectional stream handler
    service/              # business logic: AgentService, JobService, OrgService, UserService
    repository/           # GORM-backed persistence implementations
    platform/             # db connection, config, shared chi middleware
  migrations/             # goose SQL migration files
```

---

## 5. AI Directives (System Rules)

* **MUST:** Entry point is `cmd/hub/main.go`. Build target is `./cmd/hub`.
* **MUST:** HTTP routing and WebSocket handling live in `internal/transport/httpapi/`. gRPC stream handling lives in `internal/transport/grpcapi/`. Never mix transport concerns.
* **MUST:** Business logic lives in `internal/service/`. Transport handlers call services; services call repositories. Never call repository methods directly from a transport handler.
* **MUST:** Use `db.CreateInBatches(records, 500)` for any bulk insert (especially job log chunks). Never use `db.Create` inside a loop.
* **MUST:** All schema changes go through a goose migration file in `migrations/`. Never call `db.AutoMigrate` outside of a local dev helper.
* **MUST NOT:** Never import `transport/` packages from `service/` or `repository/`. Dependencies flow inward only.
* **REFERENCE:** See `docs/adr/0002-inter-service-communication.md` for protocol decisions and `docs/adr/0004-multi-tenancy-model.md` for the org-scoped data model.
