# Harmost Agent

A CLI daemon that installs itself as a system service and maintains a persistent bidirectional gRPC stream with the hub.

## Commands

### `agent run`

Runs the agent in the foreground. Blocks until Ctrl+C (interactive) or until the OS stops the service (when invoked by the service manager). This is the entry point the service manager calls internally — you typically don't invoke it directly.

```bash
nx run agent:serve
```

### `agent pair`

Pairs this agent with the hub using the OAuth2 device flow. Run this once after installing to authorize the agent.

```bash
nx run agent:pair
```

### `agent install`

Registers the agent as a system service so it starts automatically on boot. Requires elevated privileges (run as root / Administrator).

> **Note:** The service manager registers the path of the running executable. Build first (`nx run agent:build`) so the service points to a stable binary, then run with sudo.

```bash
nx run agent:build
sudo nx run agent:install
```

### `agent uninstall`

Removes the agent from the system service manager.

```bash
sudo nx run agent:uninstall
```

### `agent start`

Tells the service manager to start the already-installed service.

```bash
sudo nx run agent:start
```

### `agent stop`

Tells the service manager to stop the running service.

```bash
sudo nx run agent:stop
```

## Typical setup flow

```bash
# 1. Build (required so the service registers a stable binary path)
nx run agent:build

# 2. Install as a service (needs sudo/admin)
sudo nx run agent:install

# 3. Pair with the hub
nx run agent:pair

# 4. Start the service
sudo nx run agent:start
```

## Internal packages

| Package | Responsibility |
|---------|---------------|
| `internal/config` | Load/save config file (`hub_addr`, `grpc_addr`, `token`) |
| `internal/daemon` | `service.Interface` implementation — Start/Stop and backoff reconnect loop |
| `internal/grpc` | gRPC client — dial, hello handshake, heartbeat, hub message dispatch |
| `internal/metrics` | System metric collection (CPU, memory, disk) via gopsutil |

## Development

```bash
nx run agent:serve   # go run ./cmd/agent (foreground)
nx run agent:test    # go test ./...
nx run agent:lint    # go vet ./...
```
