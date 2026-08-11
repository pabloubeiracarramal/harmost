---
id: "0010"
title: OpenAPI Contract as the Single Source of Truth for front ↔ hub
date: 2026-08-08
status: Accepted
tags: [architecture, frontend, golang, openapi, codegen]
---

# ADR 0010: OpenAPI Contract as the Single Source of Truth for front ↔ hub

## 1. Context and Problem Statement

ADR 0002 gave hub ↔ agent a machine-readable contract: `libs/harmost-proto/proto/harmost/v1/agent.proto` generates Go for both sides, so a wire change that only lands on one end fails to compile. The front ↔ hub boundary had no equivalent. Every wire shape was written twice — a Go DTO in `internal/transport/httpapi/` and a hand-maintained TypeScript interface in `apps/front/src/features/*/api/types.ts` — and kept in agreement by discipline alone. The 16 routes themselves existed only as chi registrations and raw URL strings in the front's query hooks.

The drift this invites was already visible. `shared/ws/wsClient.ts` typed a job-status `payload.state` as bare `string` rather than the `JobState` union, because ADR 0008's layering forbids `shared/` importing from `features/`; two `as JobState` casts in `features/jobs/api/socket.ts` papered over the gap, and `apps/front/CLAUDE.md` documented the whole thing as a known wart. Separately, the front's hand-written `JobState` union was missing the `unspecified` member the hub can actually emit — a latent bug nobody had noticed.

There is a second problem this ADR must settle. ADR 0002 lists "REST + polling" as a *rejected* option for front ↔ hub and its decision line reads only "front ↔ hub: WebSocket". The system as built is REST for commands **plus** WebSocket for live events, which `CLAUDE.md`, `README.md` and `architecture.md` all describe, and which `docs/status.md` records as a documentation correction on 2026-07-13. That arrangement has never been ratified.

## 2. Considered Options

* **Option A: OpenAPI 3.1 spec as the source of truth.** One `openapi.yaml` describes the REST surface; the WS event payloads live in the same document under `components/schemas` with no `paths` entry. `oapi-codegen` generates Go models, `openapi-typescript` generates TS types.
* **Option B: Connect-RPC over the existing protobuf toolchain.** Add a `HubService` to `libs/harmost-proto` and generate Go handlers plus a typed TS client with `buf`. Literal parity with hub ↔ agent, and a server-streaming RPC could replace `/ws` outright.
* **Option C: TypeSpec.** One `.tsp` source emitting OpenAPI, protobuf and JSON Schema — the only tool that could unify all three of the project's contracts.
* **Option D: Go-first generation (Huma or swaggo).** Annotate or restructure the handlers, emit OpenAPI as a build output, then generate TS from it.

## 3. Decision

**Option A.** `libs/harmost-api/openapi.yaml` is the single source of truth for the front ↔ hub boundary, mirroring the role `libs/harmost-proto` plays for hub ↔ agent. `nx run harmost-api:generate` produces Go models (`libs/harmost-api/gen/api.gen.go`) and TypeScript types (`apps/front/src/shared/api/schema.d.ts`); both are committed, as `libs/harmost-proto/gen/` already is.

We also **ratify REST + WebSocket** for front ↔ hub: REST (`/api/v1/*`, JWT bearer) for commands and snapshots, WebSocket (`/ws?token=<jwt>`) for live push. This amends ADR 0002's decision line to match what was built.

Generation is deliberately scoped:

- **Go: models only.** No `chi-server`, no `strict-server`. The handlers keep their current shape and only their DTOs become generated types.
- **TypeScript: types only.** `shared/api/httpClient.ts` and every TanStack Query hook are untouched; each feature's `api/types.ts` becomes a re-export of the generated schema, preserving the per-feature file ADR 0008 requires.

**Justification:** Option A describes the system that already exists rather than replacing it, and can be adopted endpoint by endpoint. Option B was the closest thing to "the same as the agent contract" and is genuinely the strongest guarantee, but it means rewriting both transport layers and giving up REST URLs — too large a bet to take immediately before M5's self-hosted deploy, and browser streaming through Cloudflare Tunnel would need proving out first. Option C cannot replace `agent.proto`, because its protobuf emitter does not support streaming RPCs, so it would add a compiler without removing a contract. Option D makes the spec a build output of the hub, inverting the dependency so the front's types cannot be generated without compiling Go first.

Models-only Go generation was chosen over `strict-server` because the compile-time value is concentrated in the payload shapes — the routing table is small, stable, and already exercised end to end. `strict-server` remains available later without changing the spec.

## 4. Consequences

**Positive:**

* A wire shape now exists in exactly one place. Changing `openapi.yaml` and regenerating updates both languages; forgetting to regenerate fails `nx run harmost-api:check`.
* `shared/ws/wsClient.ts` gets its event types from a generated file rather than a feature, so it can name `JobState` without violating ADR 0008's `shared/ → features/` rule. Both `as JobState` casts are gone.
* Generating `JobState` from the hub's real enum surfaced the missing `unspecified` member, which `JobStateBadge` now renders.
* Behaviours that only existed as folklore are written down: 307 (not 302) redirects, cross-tenant access reported as 404 with no 403 anywhere, `GET /api/v1/me` answering 401 for a missing user, `/ws` being the one endpoint whose 401 is `text/plain`, and the `/ws` subscriber buffer dropping events for slow consumers.
* `oapi-codegen` is installed via a `go.mod` `tool` directive, so it needs no global install — unlike `buf`, `air` and `goose`.
* `front:typecheck` now exists. It was documented in `apps/front/CLAUDE.md` and ADR 0001 but had never been defined, so nothing on the TS side caught type drift at all.

**Negative / Trade-offs:**

* `JobSpec` now has three Go representations: proto, `domain.JobSpec` (persisted via `gorm:"serializer:json"`), and `api.JobSpec`. Letting `domain` import the generated package would invert the transport → service → repository layering, so `httpapi/convert.go` maps between the latter two. The spec marks `JobSpec`'s optional scalars `x-go-type-skip-optional-pointer` so the two structs stay field-for-field identical and the mapper stays mechanical.
* The generated `HubEvent` `oneOf` renders in Go as a `json.RawMessage` wrapper, so the four event schemas are excluded from Go generation and `events.Event` stays hand-written. Only the front consumes the union; the hub consumes the payload schemas beneath it.
* Generated output is committed, so a stale regeneration is possible between edits. `harmost-api:check` exists to catch it but there is no CI to run it yet — M5 owns that.
* Two known response quirks are documented rather than fixed, to keep this change wire-invisible: `POST /api/v1/device/token`'s 202 branch writes its body without setting `Content-Type` (so it sniffs as `text/plain`), and `POST /api/v1/jobs` / `/cancel` call `WriteHeader` before `jsonOK` sets the header.
* `features/agents/api/socket.ts` narrows with `e.type.startsWith('agent.')`, which does not narrow the union in TypeScript — it compiles only because `agent_id` happens to exist on all three event members. A future event without `agent_id` will break it.

---

## 5. AI Directives (System Rules)

* **MUST:** Any change to a front ↔ hub wire shape — a new field, a new endpoint, a new WS event — MUST be made in `libs/harmost-api/openapi.yaml` and applied by running `nx run harmost-api:generate`.
* **MUST NOT:** Never hand-edit `libs/harmost-api/gen/api.gen.go` or `apps/front/src/shared/api/schema.d.ts`. Never reintroduce a hand-written wire type in `apps/hub/internal/transport/httpapi/` or `apps/front/src/features/*/api/types.ts` — those `types.ts` files re-export generated schemas and nothing else.
* **MUST:** Keep hub ↔ agent in `libs/harmost-proto`. ADR 0002's directive is unchanged: never describe a hub ↔ agent message in `openapi.yaml`.
* **MUST:** After changing the spec, run `nx run harmost-api:check`, `nx run hub:build` and `nx run front:typecheck`. The first two catch Go drift; the third is the only thing that catches TypeScript drift.
* **REFERENCE:** See `libs/harmost-api/openapi.yaml` for the contract, ADR 0002 for the transport decisions it formalises, and ADR 0008 for the front-end boundary rules the generated types respect.
