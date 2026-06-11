# Project Status

> Keep this file up to date. Update it in the same turn as any implementation work.

## Current Focus

<!-- What is the team actively working on right now? -->
_Nothing in progress._

## Recent Changes

| Date | App | Summary |
|------|-----|---------|
| 2026-06-11 | hub | Hub structure and tooling decided — layered arch, GORM, goose, chi, buf, air; wrote ADR 0005, scaffolded package skeleton |
| 2026-06-11 | docs | Multi-tenancy model decided — org as tenant, personal org auto-created on signup; wrote ADR 0004 |
| 2026-06-06 | docs | Architecture grilling session — decided core purpose, protocols, DB, auth; wrote ADR 0002 and ADR 0003 |
| 2026-06-06 | front | Add Tailwind v4, shadcn/ui (new-york), TanStack Router (file-based), TanStack Query |
| 2026-06-06 | repo | Initial monorepo scaffolding with Nx, `agent` (Go), `hub` (Go), `front` (React/TS) |

## Pending / Backlog

<!-- Unstarted work that is planned but not yet in progress -->
- [x] **Write ADR for multi-tenancy model** — org = tenant, personal org on signup (ADR 0004)
- [ ] Implement gRPC proto file for hub ↔ agent contract
- [ ] Implement WebSocket server in hub and client in front
- [ ] Add PostgreSQL to hub (agent registry, jobs, logs, users)

## Known Issues / Blockers

<!-- Active bugs or blockers. Remove when resolved. -->
_None._
