# Hub App — AI Context

## What this app does
Cloud-hosted SaaS backend. Serves the React frontend over WebSocket and dispatches jobs to agents over gRPC bidirectional streaming. Persists agent registry, jobs, logs, users, and orgs in PostgreSQL.

## Runtime
- Language: **Go** (`github.com/harmost/hub`, Go 1.26)
- Entry point: `apps/hub/cmd/hub/main.go`

## Nx Targets

| Target | Command |
|--------|---------|
| `nx run hub:build` | Compiles to `dist/hub` |
| `nx run hub:serve` | `go run ./cmd/hub` (dev, use air instead for live reload) |
| `nx run hub:test` | `go test ./...` |
| `nx run hub:lint` | `go vet ./...` |

**MUST** use `nx run hub:<target>` — never call `go` directly.

## Package Layout

```
cmd/hub/main.go          # entry point, wires everything
internal/
  domain/                # shared types: Agent, Job, Org, User, JobLog
  transport/
    httpapi/             # chi router, REST endpoints, WebSocket, GitHub OAuth, device flow
    grpcapi/             # gRPC server, bidirectional stream handler
  service/               # business logic (AgentService, JobService, ...)
  repository/            # GORM-backed persistence
  platform/              # db connection, config, middleware
migrations/              # goose SQL migration files
```

## Tooling

| Tool | Purpose |
|------|---------|
| chi | HTTP router |
| GORM | ORM (`gorm.io/gorm` + `gorm.io/driver/postgres`) |
| goose | DB migrations |
| buf | Proto generation (hub ↔ agent contract) |
| oapi-codegen | Go model generation from `libs/harmost-api/openapi.yaml` (hub ↔ front contract) |
| air | Live reload in dev |
| testify | Test assertions |

## Conventions
- Transport handlers call services. Services call repositories. Never skip layers.
- REST request/response types are **generated** into `github.com/harmost/api/gen` (imported as `api`) from `libs/harmost-api/openapi.yaml` — see ADR 0010. Never hand-write a DTO in `httpapi/`; add the field to the spec and run `nx run harmost-api:generate`. `httpapi/convert.go` maps between `api.JobSpec` and `domain.JobSpec`, which stay separate so `domain` never imports the transport contract.
- WS event payloads are generated too, re-exported through `internal/events` (`events.JobStatusPayload`, `events.LogLine`, ...) so publishers depend on the bus's vocabulary rather than importing `api` directly. The `events.Event` envelope itself is hand-written — it carries `OrgID` as a routing key that never goes on the wire.
- Bulk inserts (especially job log chunks) **must** use `db.CreateInBatches(records, 500)`. Never `Create` inside a loop.
- Schema changes go through a goose migration file. **Never** call `AutoMigrate` outside a local dev helper.
- All resources carry `OrgID` — see ADR 0004.

## Known Gotchas
- GORM errors are returned normally but watch for silent N+1 queries on associations — use `Preload` explicitly, never rely on lazy loading.

## Connections to Other Apps
- **front** sends commands via REST (`/api/v1/*`, JWT bearer) and receives live updates via WebSocket (`/ws?token=<jwt>`).
- **agent** connects via gRPC bidirectional stream (agent always initiates).
- See [architecture overview](../../docs/architecture.md) and [ADR 0002](../../docs/adr/0002-inter-service-communication.md).
