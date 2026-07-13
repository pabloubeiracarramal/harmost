# Project Status

> Keep this file up to date. Update it in the same turn as any implementation work.

## Current Focus

MVP push — see [docs/roadmap.md](roadmap.md). M1 (agent Docker executor, #16/#17) implemented and verified E2E on 2026-07-13: hub REST dispatch → gRPC stream → agent container run → status/logs persisted (success, non-zero exit, container cleanup all confirmed). Next: M2.

## Recent Changes

| Date | App | Summary |
|------|-----|---------|
| 2026-07-13 | agent | Docker executor (M1) — run.go full container lifecycle (pull per policy → create → start → log streaming with stdout/stderr demux → wait, force-remove cleanup); manager.go Dispatch/Cancel with per-job timeout ctx and terminal JobState mapping |
| 2026-07-13 | agent | gRPC job wiring — DispatchJob/CancelJob handled, job status/log messages funneled through send channel (single-sender stream), Ping now answered with Pong; daemon creates Docker+Manager at boot (degrades gracefully without Docker) |
| 2026-07-13 | agent | `agent containers` debug command — lists all host containers with their harmost job ID; runnable via `nx run agent:containers` |
| 2026-07-13 | agent | internal/docker started (M1) — moby SDK wrapper (New/Ping/ListAllContainers) + JobSpec→container.Config translation in spec.go |
| 2026-07-08 | docs | docs/roadmap.md — 8-week MVP plan; GitHub tracker reconciled (12 stale issues closed, milestones M1–M6, issues #25–#33) |
| 2026-07-08 | repo | Dev toolset — compose.dev.yaml (Postgres+Adminer), workspace Nx targets (db, db:reset, dev, grpc:ui), apps/hub/api.http, docs/dev.md, root .envrc; installed grpcurl/grpcui/dlv |
| 2026-06-19 | hub | Agent metrics — migration 009 adds metrics columns to agents; heartbeats persist CPU/mem/disk snapshots, exposed via REST |
| 2026-06-19 | hub | Migration 008 — agent_tokens.agent_id binds pairing tokens to their agent |
| 2026-06-19 | agent | Host metrics via gopsutil — CPU, memory, disk collected and sent in heartbeats |
| 2026-06-19 | front | /agents/$id — agent detail page with live metric gauges (WebSocket) |
| 2026-06-13 | hub | Migration 007 — alter users (github_id, name, avatar_url), add agent_tokens + device_codes tables |
| 2026-06-13 | hub | GitHub OAuth — /auth/github + /auth/github/callback; issues signed JWT; redirects to frontend |
| 2026-06-13 | hub | JWT auth middleware replaces x-org-id placeholder on all protected HTTP routes |
| 2026-06-13 | hub | Agent token auth — gRPC Connect validates `authorization: Bearer <token>` against agent_tokens table |
| 2026-06-13 | hub | Device flow — POST /api/v1/device/authorize, POST /api/v1/device/token, POST /api/v1/device/approve |
| 2026-06-13 | hub | WebSocket /ws?token=<jwt> — streams agent.connected/disconnected/heartbeat events to frontend clients |
| 2026-06-13 | hub | internal/auth/jwt.go + internal/events/bus.go — JWT helpers and pub/sub event bus |
| 2026-06-13 | agent | pair command — device flow: POST authorize, print URL, poll token endpoint, persist config |
| 2026-06-13 | agent | daemon — gRPC connect loop with Bearer token auth, AgentHello, 30s heartbeats, exponential backoff |
| 2026-06-13 | agent | internal/config — OS-appropriate config.json load/save |
| 2026-06-13 | front | Vite proxy → hub:8080 for /auth, /api, /ws |
| 2026-06-13 | front | /login — GitHub OAuth button; /auth/callback — stores JWT, redirects |
| 2026-06-13 | front | /dashboard — agent list (REST) + live status updates (WebSocket) |
| 2026-06-13 | front | /device?code=XXXX — device approval page |
| 2026-06-12 | hub | main.go wired — gRPC (:50051) + HTTP (:8080) servers, graceful shutdown |
| 2026-06-12 | hub | HTTP transport — chi router, REST endpoints for agents + jobs |
| 2026-06-12 | hub | gRPC transport — AgentService.Connect bidirectional stream; in-memory stream registry |
| 2026-06-12 | hub | Service + repository layers — all domain entities backed by GORM |
| 2026-06-12 | hub | Domain models + 6 goose migrations — users, orgs, org_members, agents, jobs, job_logs |
| 2026-06-12 | proto | gRPC contract — AgentService.Connect bidi stream; shared go.work module |
| 2026-06-11 | hub | Scaffolded hub structure — layered arch, tooling, ADR 0005 |
| 2026-06-11 | docs | Multi-tenancy model — ADR 0004 |
| 2026-06-06 | docs | Architecture decisions — ADR 0002, ADR 0003 |
| 2026-06-06 | front | Tailwind v4, shadcn/ui, TanStack Router + Query |
| 2026-06-06 | repo | Initial Nx monorepo scaffolding |

## Pending / Backlog

- [ ] WebSocket auto-reconnect on the frontend (currently closes on error)
- [ ] JWT refresh / token expiry handling in frontend (tokens expire in 24h)
- [ ] GET /api/v1/me — return current user profile
- [ ] Job update loss on disconnect — status/log sends are dropped once the 256-message buffer fills while the stream is down; needs hub-side reconciliation on reconnect
- [ ] JobSpec HostConfig mapping — volume mounts, resource limits, network mode, privileged not yet translated (post-MVP per spec.go)
- [ ] Multi-org support — org switcher; users currently always use their personal org

## Known Issues / Blockers

_None currently. (Docker Desktop WSL integration was enabled on 2026-07-13, unblocking local Postgres and the agent's Docker executor.)_
