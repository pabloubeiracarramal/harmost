# Agent App — AI Context

## What this app does
A CLI tool that installs itself as a system service (via `kardianos/service`), maintains a persistent bidirectional gRPC stream with the hub, and executes dispatched jobs as Docker containers (moby SDK), streaming status and logs back up the stream. Pairing to a hub is done via OAuth2 device flow (`pair` command).

## Runtime
- Language: **Go** (`github.com/harmost/agent`, Go 1.26)
- Entry point: `apps/agent/main.go`
- CLI framework: **cobra**
- Service manager: **kardianos/service** (cross-platform daemon install/uninstall/start/stop)
- Hub transport: **gRPC** bidirectional streaming
- Auth: OAuth2 device flow for the `pair` command

## CLI Commands

| Command | Description |
|---------|-------------|
| `agent install` | Install the agent as a system service |
| `agent uninstall` | Uninstall the system service |
| `agent start` | Start the installed service |
| `agent stop` | Stop the running service |
| `agent run` | Run the agent in the foreground (called by the service manager) |
| `agent pair` | Pair this agent to a hub using OAuth2 device flow |
| `agent containers` | List all containers on the host with their harmost job ID (debugging) |

## Nx Targets

| Target | Command |
|--------|---------|
| `nx run agent:build` | Compiles to `dist/agent` |
| `nx run agent:serve` | `go run ./cmd/agent run` (foreground dev) |
| `nx run agent:test` | `go test ./...` |
| `nx run agent:lint` | `go vet ./...` |
| `nx run agent:containers` | `go run ./cmd/agent containers` (debugging) |

**MUST** use `nx run agent:<target>` — never call `go` directly.

## Key Design Points
- `kardianos/service` drives the service lifecycle. The cobra `run` command is the service entry point passed to `service.New(...)`.
- The `pair` command opens the device flow URL, polls the token endpoint, and persists credentials locally for use by the service.
- The gRPC stream to hub is bidirectional (`bidi stream`): hub pushes tasks/config down, agent streams results/events up. The stream is reconnected with exponential back-off on disconnect.
- Credentials and hub address are stored in a config file (location is OS-appropriate via `os.UserConfigDir`).

## Conventions
- Standard Go project layout; entry at package root (`main.go`).
- Tests live alongside source files (`_test.go`).
- No ORM — use standard library or explicit SQL drivers if a DB is added.

## Known Gotchas
<!-- Add sharp edges, non-obvious invariants, or past bugs here -->
- gRPC streams allow only ONE sending goroutine. All async sends (job status, logs, pong) must go through the `sendCh` channel in `internal/grpc/client.go` — never call `stream.Send` from another goroutine.
- Jobs are bound to the daemon's root context, not the stream's: they survive hub reconnects. Messages sent while the stream is down are dropped once the 256-message buffer fills.
- `docker.Run` returns `(exitCode, nil)` for a non-zero container exit — a non-nil error means the lifecycle itself failed. Terminal JobStates are decided in `manager.go`, never in `run.go`.

## Connections to Other Apps
- Connects to: **hub** — bidirectional gRPC stream; see [architecture overview](../../docs/architecture.md) for protocol details.
- Pairs with: **hub** OAuth2 device flow endpoint.

## Folder Structure
apps/agent/
├── cmd/
│   └── agent/
│       ├── main.go         # Bootstraps the app
│       ├── root.go         # Defines rootCmd and service helper
│       ├── run.go          # 'agent run'
│       ├── install.go      # 'agent install'
│       ├── uninstall.go    # 'agent uninstall'
│       ├── start.go        # 'agent start'
│       ├── stop.go         # 'agent stop'
│       ├── pair.go         # 'agent pair' — OAuth2 device flow
│       └── containers.go   # 'agent containers' — list host containers (debugging)
└── internal/
    ├── config/
    │   └── config.go       # Load/Save config (hub addr, token)
    ├── daemon/
    │   └── program.go      # service.Interface: Start/Stop + backoff reconnect loop
    ├── docker/
    │   ├── docker.go       # Moby SDK client wrapper (New, Ping, ListAllContainers)
    │   ├── spec.go         # proto JobSpec → container.Config/HostConfig translation
    │   ├── run.go          # One job's lifecycle: pull → create → start → stream logs → wait
    │   └── manager.go      # Dispatch/Cancel — tracks running jobs, maps terminal JobStates
    ├── grpc/
    │   └── client.go       # Dial, hello handshake, heartbeat loop, send channel, job dispatch
    └── metrics/
        └── metrics.go      # System metric collection (CPU, memory, disk)
