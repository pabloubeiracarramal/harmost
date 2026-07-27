---
id: "0007"
title: Hub Runs as a Single Instance (No Horizontal Scaling)
date: 2026-07-21
status: Accepted
tags: [architecture, golang, scalability, hub]
---

# ADR 0007: Hub Runs as a Single Instance (No Horizontal Scaling)

## 1. Context and Problem Statement

The hub holds two pieces of state entirely in process memory:

- `internal/events/bus.go` — the WS event bus (`Bus.subs`), a plain `map[string]map[int]chan Event` keyed by org ID, local to one process's heap.
- `internal/transport/grpcapi/grpcapi.go` — the agent stream registry (`registry.streams`), a `map[string]sendFn` keyed by agent ID, where each `sendFn` is a closure bound to one specific gRPC stream on one specific process.

Neither is shared across processes. If the hub is ever run as more than one concurrent replica behind a load balancer, two things break silently:

- A REST dispatch/cancel request routed to a replica that doesn't hold the target agent's gRPC stream sees `Connected(agentID) == false` and rejects it (404/409), even though the agent is online on a different replica.
- A WebSocket client connected to a replica that isn't publishing an event (e.g. a job status change surfaced by an agent stream held on another replica) never receives it — the UI goes quiet with no error.

This was raised during an architecture review. At current scale (solo-dev MVP, M4/M5 in flight) a single hub process is sufficient — restarts are already handled reasonably (agents reconnect with backoff, the orphan sweeper fails stale jobs, REST backfill covers the WS gap on reconnect). The risk is someone deploying multiple replicas later, expecting failover/load-sharing, and getting silently broken dispatch and dead WebSocket updates instead of a clear error.

## 2. Considered Options

* Option A: Externalize both stores now — event bus over Postgres `LISTEN/NOTIFY` or Redis pub/sub, plus a shared agent→instance directory and inter-hub forwarding for dispatch/cancel, so multiple replicas work correctly.
* Option B: Keep the hub single-instance by design for now; explicitly document the constraint so it isn't hit by surprise, and revisit Option A only if/when load or availability requirements actually demand more than one replica.

## 3. Decision

**Option B.** The hub is single-instance by design until a concrete need for horizontal scaling arises.

**Justification:** Option A is a real chunk of work — a new pub/sub dependency or LISTEN/NOTIFY plumbing, plus a connection-broker layer (shared directory + inter-hub RPC) for the gRPC registry, which is the harder half since a gRPC stream cannot be handed to another process. At current solo-dev MVP scale, vertical scaling of one instance covers substantial headroom, and building the multi-replica machinery now would be speculative — there's no current load or HA requirement driving it. Recording the constraint explicitly is cheap insurance against a future deploy misconfiguration (e.g. someone bumping replica count on Fly.io/Railway during M5) causing silent breakage instead of a clear failure.

## 4. Consequences

**Positive:**
* No new infrastructure dependency (Redis, etc.) or connection-broker complexity added before it's needed.
* Deployment stays simple: one hub process, one gRPC port, one HTTP port.
* The constraint is now written down, so M5 deployment work (Fly.io/Railway) won't accidentally configure multiple replicas expecting them to share agent connections or WS events.

**Negative / Trade-offs:**
* No horizontal scaling or multi-replica HA for the hub. A hub process restart/crash drops every agent connection at once (mitigated by agent reconnect-with-backoff and the orphan sweeper, but still a single point of failure for the whole fleet during the outage window).
* If load or availability needs later require more than one replica, Option A's work (external pub/sub + agent-instance directory + inter-hub forwarding) still has to be done — this ADR defers it, it doesn't remove it.

---

## 5. AI Directives (System Rules)

* **MUST:** Deploy and configure the hub as exactly one running instance. Do not add a second concurrent hub replica behind a load balancer without first implementing Option A (externalized event bus + shared agent registry/connection broker).
* **MUST NOT:** Do not "fix" cross-replica dispatch or WS delivery gaps with partial patches (e.g. sticky sessions at the LB) — the gRPC agent registry and WS event bus both need a real shared backing store; partial fixes will mask the problem instead of solving it.
* **REFERENCE:** See `docs/architecture.md` Topology section and `apps/hub/internal/events/bus.go` / `apps/hub/internal/transport/grpcapi/grpcapi.go` for the current in-memory implementations this ADR constrains.
