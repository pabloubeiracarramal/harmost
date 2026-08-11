# harmost-proto

Protobuf contract for **hub ↔ agent**. `proto/harmost/v1/agent.proto` defines `AgentService`, a single
bidirectional gRPC stream (`Connect`) that the agent initiates and keeps open — the hub pushes job
dispatch/cancel/ping over it, the agent pushes hello/status/logs/heartbeat/pong back. `buf` generates
Go for both sides from that one file; there is no hand-written DTO on either end of this boundary.

See ADR 0002 (`docs/adr/0002-inter-service-communication.md`) for why gRPC was chosen here, and
`docs/architecture.md` for how this fits the rest of the system.

## Layout

```
buf.yaml           # module + lint/breaking-change rules (STANDARD lint, FILE-level breaking checks)
buf.gen.yaml        # codegen plugins: protoc-gen-go, protoc-gen-go-grpc (both via buf.build remote)
proto/harmost/v1/agent.proto   # the contract — the only file you hand-edit
gen/                # generated Go — committed, never hand-edited
```

## Editing the contract

1. Edit `proto/harmost/v1/agent.proto`. The two `oneof` envelopes, `HubMessage` (hub → agent) and
   `AgentMessage` (agent → hub), are the extension points:
   - **New hub → agent message:** add a message type, add it as a new field in `HubMessage.payload`.
   - **New agent → hub message:** add a message type, add it as a new field in `AgentMessage.payload`.
   - Field numbers are wire identity — never reuse or renumber an existing field; only append.
2. Regenerate:
   ```
   nx run harmost-proto:generate
   ```
   Runs `buf generate`, emitting into `gen/` (Go structs + gRPC client/server stubs). Commit the
   result alongside the `.proto` change.
3. Lint the contract itself:
   ```
   nx run harmost-proto:lint
   ```
   `buf lint` (STANDARD ruleset, with `RPC_REQUEST_STANDARD_NAME`/`RPC_RESPONSE_STANDARD_NAME`
   excepted) plus `buf breaking` (FILE-level) to catch accidental wire breaks.
4. Wire it up by hand on both sides — generation only gives you types and stream plumbing, not
   business logic:
   - Hub: `apps/hub/internal/transport/grpcapi/connect.go` (dispatches on the incoming oneof).
   - Agent: `apps/agent/internal/grpc/client.go` (`handleMessage` and friends).
5. Build both to catch drift:
   ```
   nx run hub:build
   nx run agent:build
   ```

## Rules

- Never hand-edit anything under `gen/` — regenerate instead.
- Never describe a front ↔ hub message here. That boundary has its own contract,
  `libs/harmost-api` (OpenAPI) — see ADR 0010 (`docs/adr/0010-openapi-contract-for-front-hub.md`).
- `buf` is a global binary (`$(go env GOPATH)/bin/buf`), unlike `harmost-api`'s `oapi-codegen`, which
  installs via a `go.mod` tool directive — no global install needed there.
