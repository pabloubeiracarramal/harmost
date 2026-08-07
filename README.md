# Harmost

CI/CD orchestration SaaS. A cloud-hosted **hub** dispatches Docker-based jobs to **agents** installed on users' own machines; a React **front** lets users pair agents, dispatch jobs, and watch live status and logs.

```
front (React)  ──REST + WebSocket──▶  hub (Go, cloud)  ◀──gRPC bidi stream──  agent (Go, user machine)  ──▶  Docker
```

## Apps

| App | Stack | What it does |
|-----|-------|--------------|
| [`apps/front`](apps/front) | React 19 + TypeScript + Vite | Web UI — login, agent dashboard, job dispatch, live logs |
| [`apps/hub`](apps/hub) | Go | SaaS backend — REST + WebSocket for the UI, gRPC server for agents, PostgreSQL persistence |
| [`apps/agent`](apps/agent) | Go | System-service daemon — pairs via OAuth2 device flow, runs Docker jobs, streams status/logs to the hub |
| [`libs/harmost-proto`](libs/harmost-proto) | Protobuf (buf) | gRPC contract shared by hub and agent |

## Quickstart (local dev)

```sh
npm install
nx run workspace:dev     # Postgres → hub (live reload) → front (Vite on :4200)
```

Full setup (Docker, hub `.env`, running an agent, API/gRPC/WS tooling): see [docs/dev.md](docs/dev.md).

## Documentation

- [docs/architecture.md](docs/architecture.md) — system overview, data flow, boundaries
- [docs/roadmap.md](docs/roadmap.md) — 8-week MVP plan and milestones
- [docs/status.md](docs/status.md) — current focus, recent changes, backlog
- [docs/dev.md](docs/dev.md) — local development guide
- [docs/adr/](docs/adr) — architecture decision records

## Conventions

- All builds/tests go through Nx: `nx run <project>:<target>` (never `go`/`vite`/`tsc` directly).
- Conventional Commits; feature branches (`feat/...`, `fix/...`); PRs target `develop`.
