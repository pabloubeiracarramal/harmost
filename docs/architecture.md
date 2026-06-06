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
   WebSocket  (TanStack Query + WS client)
       |
  [ hub (Go) — cloud ]
       |
   gRPC bidi stream
       |
  [ agent (Go) — user machine ]
       |
   docker run / docker build
       |
  [ Docker daemon ]
```

### Job dispatch (UI-triggered)

```
front  →(WS)→  hub  →(gRPC)→  agent  →  docker run <image>
```

### Job dispatch (git webhook)

```
GitHub/GitLab webhook  →(HTTPS)→  hub  →(gRPC)→  agent  →  docker run <image>
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
| front ↔ hub | WebSocket | bidirectional | Job events, agent status, live log chunks |
| hub ↔ agent | gRPC bidirectional streaming | agent initiates | Agent is gRPC client; hub is server. Agent reconnects with exponential back-off |
| git provider → hub | HTTPS webhook | inbound | GitHub/GitLab push/PR events trigger job dispatch |
| hub ↔ PostgreSQL | TCP (pgx driver) | hub-initiated | Job history, agent registry, user accounts |

## Authentication

- **UI users**: OAuth2 / social login (GitHub, Google). Hub is the OAuth2 client; sessions stored in PostgreSQL.
- **Agent pairing**: Agent runs `agent pair` → OAuth2 device flow against hub → credentials stored in OS config directory → used by the service process on all subsequent gRPC connections.

## Agent Routing

The user (or webhook payload) explicitly names the target agent. Hub looks up the named agent's live gRPC stream and forwards the job spec. No automatic load-balancing yet.

## Topology

Many agents connect to one hub instance. Each agent represents a distinct machine or environment (e.g., `prod-server-01`, `build-arm64`).

## Persistence (hub)

PostgreSQL. Stores:
- Agent registry (id, name, last-seen, pairing credentials)
- Jobs (id, agent, trigger, status, timestamps)
- Job logs (chunked, indexed by job id)
- Users and OAuth sessions

## Shared Libraries

`libs/` is currently empty. Future candidates:
- gRPC proto definitions shared by hub and agent
- Shared TypeScript types if an API schema layer is added

Run `nx graph` to see the live dependency graph.

## Relevant ADRs

- [ADR 0001 — Nx as monorepo build tool](adr/0001-nx-monorepo-tooling.md)
- [ADR 0002 — Inter-service communication protocols](adr/0002-inter-service-communication.md)
- [ADR 0003 — Language and framework choices](adr/0003-language-and-framework-choices.md)
