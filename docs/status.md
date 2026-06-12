# Project Status

> Keep this file up to date. Update it in the same turn as any implementation work.

## Current Focus

<!-- What is the team actively working on right now? -->
_Nothing in progress._

## Recent Changes

| Date | App | Summary |
|------|-----|---------|
| 2026-06-12 | hub | main.go wired — gRPC (:50051) + HTTP (:8080) servers, graceful shutdown on SIGTERM/SIGINT |
| 2026-06-12 | hub | HTTP transport — chi router, orgID middleware, agents + jobs REST endpoints (list/get/dispatch/logs) |
| 2026-06-12 | hub | gRPC transport — AgentService.Connect bidirectional stream handler; in-memory stream registry for job dispatch; proto↔domain converters |
| 2026-06-12 | hub | Service layer — UserService (signup tx), AgentService (connect/disconnect/heartbeat), JobService (dispatch/state transitions), JobLogService (batch ingest) |
| 2026-06-12 | hub | Repository layer — UserRepo, OrgRepo (+ members), AgentRepo, JobRepo, JobLogRepo; all backed by GORM |
| 2026-06-12 | hub | Domain models, platform (config + GORM), and 6 goose migrations applied to dev DB — users, orgs, org_members, agents, jobs, job_logs |
| 2026-06-12 | proto | gRPC contract defined — `AgentService.Connect` bidirectional stream; `libs/harmost-proto` shared module with `go.work`; generated Go stubs |
| 2026-06-11 | hub | Hub structure and tooling decided — layered arch, GORM, goose, chi, buf, air; wrote ADR 0005, scaffolded package skeleton |
| 2026-06-11 | docs | Multi-tenancy model decided — org as tenant, personal org auto-created on signup; wrote ADR 0004 |
| 2026-06-06 | docs | Architecture grilling session — decided core purpose, protocols, DB, auth; wrote ADR 0002 and ADR 0003 |
| 2026-06-06 | front | Add Tailwind v4, shadcn/ui (new-york), TanStack Router (file-based), TanStack Query |
| 2026-06-06 | repo | Initial monorepo scaffolding with Nx, `agent` (Go), `hub` (Go), `front` (React/TS) |

## Pending / Backlog

- [ ] Auth — GitHub OAuth for users, org-scoped tokens for agents (ADR 0006); `x-org-id` header is a dev placeholder
- [ ] WebSocket server in hub + client in front (live job/agent updates)
- [ ] Agent app — implement gRPC client, AgentHello handshake, job runner
- [ ] Frontend — connect to hub API (agent list, job dispatch, live logs)
- [ ] User signup flow (UserService.SignUp is ready, needs an HTTP endpoint + auth provider)

## Known Issues / Blockers

<!-- Active bugs or blockers. Remove when resolved. -->
_None._
