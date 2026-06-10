# Harmost Agent

A CLI daemon that installs itself as a system service and maintains a persistent bidirectional gRPC stream with the hub.

## Building

```bash
nx run agent:build   # compiles to dist/agent
```

## Commands

### `agent run`

Runs the agent in the foreground. Blocks until Ctrl+C (interactive) or until the OS stops the service (when invoked by the service manager). This is the entry point the service manager calls internally — you typically don't invoke it directly.

```bash
nx run agent:serve
# or after building:
./dist/agent run
```

### `agent pair`

Pairs this agent with the hub using the OAuth2 device flow. Run this once after installing to authorize the agent.

```bash
./dist/agent pair
```

### `agent install`

Registers the agent as a system service so it starts automatically on boot. Requires elevated privileges (run as root / Administrator).

```bash
./dist/agent install
```

### `agent uninstall`

Removes the agent from the system service manager.

```bash
./dist/agent uninstall
```

### `agent start`

Tells the service manager to start the already-installed service.

```bash
./dist/agent start
```

### `agent stop`

Tells the service manager to stop the running service.

```bash
./dist/agent stop
```

## Typical setup flow

```bash
# 1. Build
nx run agent:build

# 2. Install as a service (needs sudo/admin)
sudo ./dist/agent install

# 3. Pair with the hub
./dist/agent pair

# 4. Start the service
sudo ./dist/agent start
```

## Development

```bash
nx run agent:serve   # go run ./... from apps/agent
nx run agent:test    # go test ./...
nx run agent:lint    # go vet ./...
```
