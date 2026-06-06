# Hub App — AI Context

## What this app does
<!-- One sentence: what is the hub's role? e.g. "Central API gateway that authenticates front requests and dispatches work to agent." -->
_TBD — fill in once the role is defined._

## Runtime
- Language: **Go** (`github.com/harmost/hub`, Go 1.26)
- Entry point: `apps/hub/` (single `./...` package tree)

## Nx Targets

| Target | Command |
|--------|---------|
| `nx run hub:build` | Compiles to `dist/hub` |
| `nx run hub:serve` | `go run ./...` (dev) |
| `nx run hub:test` | `go test ./...` |
| `nx run hub:lint` | `go vet ./...` |

**MUST** use `nx run hub:<target>` — never call `go` directly.

## Conventions
- Standard Go project layout; entry is at package root (no `cmd/` yet).
- Tests live alongside source files (`_test.go`).
- No ORM — use standard library or explicit SQL drivers if a DB is added.

## Known Gotchas
<!-- Add sharp edges, non-obvious invariants, or past bugs here -->
_None yet._

## Connections to Other Apps
- Downstream client: **front** — serves the UI's API requests.
- Downstream caller: **agent** — dispatches work to agent.
- See [architecture overview](../../docs/architecture.md) for protocol details.
