# Architecture Overview

> Update this document whenever the data flow, boundaries, or responsibilities change.
> For major structural decisions, create an ADR in `/docs/adr/`.

## Product

Harmost is a **CI/CD orchestration SaaS**. You (Harmost) host the Hub in the cloud. Users install the Agent as a system service on their own machines/servers. The React frontend lets users monitor agents and dispatch Docker-based jobs.

## Apps at a Glance

| App | Runtime | Module | Responsibility |
|-----|---------|--------|----------------|
| `apps/front` | React + TypeScript (Vite) | — | Web UI: view agents, trigger jobs, stream live logs |
| `apps/hub` | Go (`github.com/harmost/hub`) | — | Cloud SaaS backend: gRPC server for agents, WebSocket server for the frontend, job queue, PostgreSQL persistence |
| `apps/agent` | Go (`github.com/harmost/agent`) | — | System-service daemon installed on user machines; pairs via OAuth2 device flow, runs Docker jobs, streams logs back to hub |

## Data Flow

```
  User (browser)
       |
   REST (commands) + WebSocket (live events)
       |
  [ hub (Go) — cloud ]
       |
   gRPC bidi stream
       |
  [ agent (Go) — user machine ]
       |
   container run (moby SDK)
       |
  [ Docker daemon ]
```

### Job dispatch (UI-triggered)

```
front  →(REST: POST /api/v1/jobs)→  hub  →(gRPC DispatchJob)→  agent  →  Docker container
```

### Job dispatch (git webhook) — planned, not yet implemented

```
GitHub/GitLab webhook  →(HTTPS)→  hub  →(gRPC)→  agent  →  Docker container
```

### Log streaming

```
docker container stdout/stderr
   → agent tails logs
   → streams chunks over gRPC bidi stream
   → hub forwards to any WebSocket client watching that job
   → front renders live in UI
```

## Key Boundaries

| Boundary | Protocol | Direction | Notes |
|---|---|---|---|
| front → hub (commands) | REST (`/api/v1/*`, JWT bearer) | front-initiated | Agent/job CRUD, job dispatch, device-flow approval |
| hub → front (events) | WebSocket (`/ws?token=<jwt>`) | hub pushes | Agent connected/disconnected/heartbeat events (job events planned — M2) |
| hub ↔ agent | gRPC bidirectional streaming | agent initiates | Agent is gRPC client; hub is server. Bearer agent-token auth. Agent reconnects with exponential back-off |
| git provider → hub | HTTPS webhook | inbound | **Planned.** GitHub/GitLab push/PR events would trigger job dispatch |
| hub ↔ PostgreSQL | TCP (GORM, postgres driver) | hub-initiated | Job history, agent registry, user accounts |

## Authentication

See [ADR 0006](adr/0006-authentication-strategy.md).

- **UI users**: GitHub OAuth (hub is the OAuth2 client) → hub issues a signed JWT (24h) that the frontend stores and sends as a bearer token on REST and WebSocket requests. Additional providers (Google, GitLab) are post-MVP.
- **Agent pairing**: Agent runs `agent pair` → OAuth2 device flow against hub → org-scoped agent token stored in OS config directory → sent as `authorization: Bearer <token>` metadata on every gRPC connection, validated against the `agent_tokens` table.

## Agent Routing

The user (or webhook payload) explicitly names the target agent. Hub looks up the named agent's live gRPC stream and forwards the job spec. No automatic load-balancing yet.

## Topology

Many agents connect to one hub instance. Each agent represents a distinct machine or environment (e.g., `prod-server-01`, `build-arm64`).

The hub is single-instance by design (see [ADR 0007](adr/0007-hub-single-instance.md)): the WS event bus and the gRPC agent stream registry are both in-process memory, not shared across replicas. Running more than one concurrent hub process will silently break dispatch/cancel and WebSocket delivery for agents/events held on a different replica.

## Persistence (hub)

PostgreSQL (GORM; schema managed by goose migrations). Stores:
- Users, orgs, org memberships (multi-tenancy per ADR 0004)
- Agent registry (id, name, last-seen, metrics snapshots) and agent tokens
- Jobs (id, agent, status, exit code, timestamps) and device codes
- Job logs (chunked, indexed by job id)

## Shared Libraries

- `libs/harmost-proto` — the gRPC contract (`AgentService.Connect` bidi stream), managed with buf; generated Go code consumed by hub and agent via `go.work`.

Future candidate: shared TypeScript types if an API schema layer is added. Run `nx graph` to see the live dependency graph.

## Relevant ADRs

- [ADR 0001 — Nx as monorepo build tool](adr/0001-nx-monorepo-tooling.md)
- [ADR 0002 — Inter-service communication protocols](adr/0002-inter-service-communication.md)
- [ADR 0003 — Language and framework choices](adr/0003-language-and-framework-choices.md)
- [ADR 0004 — Multi-tenancy model](adr/0004-multi-tenancy-model.md)
- [ADR 0005 — Hub internal structure and tooling](adr/0005-hub-structure-and-tooling.md)
- [ADR 0006 — Authentication strategy](adr/0006-authentication-strategy.md)
- [ADR 0007 — Hub runs as a single instance (no horizontal scaling)](adr/0007-hub-single-instance.md)
