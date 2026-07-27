# Local Development Guide

How to run the stack and poke at every surface of it while you build.

## One-time setup

1. **Docker** — enable WSL integration in Docker Desktop (Settings → Resources →
   WSL Integration → toggle this distro), or install Docker Engine natively in WSL.
   Verify with `docker ps`.
2. **Hub env** — `apps/hub/.env` (copy from `.env.example` if missing). Needs a
   GitHub OAuth app (callback `http://localhost:8080/auth/github/callback`).
3. **psql (optional)** — `sudo apt-get install -y postgresql-client`. Adminer
   (browser) works without it.

## Daily workflow

```sh
nx run workspace:dev        # Postgres (+Adminer) → hub (air) → front (vite)
```

Or piece by piece:

| Command | What it does |
|---------|--------------|
| `nx run workspace:db` | Start Postgres + Adminer in Docker |
| `nx run hub:migrate` | Apply goose migrations (needs `DATABASE_URL` in env) |
| `nx run hub:dev` | Hub with air live-reload (HTTP :8080, gRPC :50051) |
| `nx run front:dev` | Vite dev server on :4200 (proxies /auth, /api, /ws to hub) |
| `nx run workspace:db:reset` | Drop the DB volume, recreate, re-migrate |
| `nx run workspace:db:down` | Stop the DB containers |

Run the agent against the local hub:

```sh
nx run agent:build
./apps/agent/dist/agent pair --hub http://localhost:8080   # once
./apps/agent/dist/agent run                                # foreground daemon
```

## Trying things as you build

### REST API
Open `apps/hub/api.http` in VS Code (REST Client extension) — every endpoint has
a ready-made request. Get a JWT by logging in at `http://localhost:4200/login`
and copying the `?token=` value from the callback URL (or from localStorage).
Headless alternative: `go run ./cmd/devtoken <user-id> <org-id>` from `apps/hub/`
(direnv loads `JWT_SECRET`); look up the IDs in the `users`/`orgs` tables.

### Database
- **Browser:** Adminer at http://localhost:8081 (server `db`, user/pass `postgres`).
- **Terminal:** `psql "$DATABASE_URL"` from `apps/hub/` (direnv loads `.env`).

### gRPC (the hub↔agent stream)
```sh
nx run workspace:grpc:ui    # browser UI for AgentService at :50051
```
grpcui/grpcurl load the proto from `libs/harmost-proto/proto`. The `Connect`
stream needs metadata `authorization: Bearer <agent-token>` — add it in the
grpcui "Metadata" section. Get a token via the device-flow requests in `api.http`.

```sh
# CLI equivalent
grpcurl -plaintext -import-path libs/harmost-proto/proto \
  -proto harmost/v1/agent.proto localhost:50051 list
```

### WebSocket (the hub→front event stream)
```sh
wscat -c "ws://localhost:8080/ws?token=<jwt>"
```
Then connect/disconnect an agent and watch `agent.connected` / `agent.heartbeat`
events arrive.

### Debugger
`dlv` is installed. In VS Code, a Go launch config pointed at
`apps/hub/cmd/hub` (or attach to the air-built `tmp/hub` process) gives you
breakpoints. Remember the hub reads config from env — direnv in `apps/hub/`
loads `.env` for terminal launches; VS Code needs `"envFile"` in the launch config.

### Code navigation
`codegraph explore "<topic>"` / `codegraph node <symbol|file>` — call paths,
blast radius, and source for any symbol (index auto-syncs).

### gRPC TLS — local dev vs production

The hub serves gRPC in **plaintext by default** (matches local dev — no cert
needed). Production must configure TLS:

| Mode | Hub | Agent |
|------|-----|-------|
| Local dev (default) | No `GRPC_TLS_*` env set → plaintext | `agent pair <hub-url> --insecure` persists `insecure: true` in `config.json` |
| Production | Set `GRPC_TLS_CERT_FILE` + `GRPC_TLS_KEY_FILE` (or terminate TLS in a proxy in front of `:50051` and leave these unset) | `agent pair <hub-url>` (no `--insecure`) — dials with the system cert pool |

The hub refuses to start if only one of `GRPC_TLS_CERT_FILE`/`GRPC_TLS_KEY_FILE`
is set, and logs a warning if it serves plaintext with `ENV=production`. The
`--insecure` choice is baked into the agent's persisted config at pair time —
re-pair to switch modes.

## Ports

| Port | Service |
|------|---------|
| 4200 | front (Vite) |
| 8080 | hub HTTP + WebSocket |
| 50051 | hub gRPC |
| 5432 | Postgres |
| 8081 | Adminer |
