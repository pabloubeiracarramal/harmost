
## Project Overview

Harmost is a CI/CD orchestration SaaS: a cloud-hosted **hub** (Go) dispatches Docker-based jobs to **agents** (Go daemons) installed on users' machines, monitored and controlled from a React **front**. Front ↔ hub is REST + WebSocket; hub ↔ agent is a gRPC bidirectional stream (agent initiates). See `docs/architecture.md` for the full picture and `docs/status.md` for current state.

## Monorepo reference

| Path | Contents |
|------|----------|
| `apps/front` | React 19 + TS + Vite web UI (see `apps/front/CLAUDE.md`) |
| `apps/hub` | Go backend — REST/WS + gRPC server + PostgreSQL (see `apps/hub/CLAUDE.md`) |
| `apps/agent` | Go CLI/daemon — Docker job executor (see `apps/agent/CLAUDE.md`) |
| `libs/harmost-proto` | Protobuf contract (buf) shared by hub and agent |
| `docs/` | architecture, roadmap, status, dev guide, ADRs |

Everything runs through Nx: `nx run <project>:<target>`. Projects: `front`, `hub`, `agent`, `workspace` (root targets: `dev`, `db`, `db:reset`, `grpc:ui`).


## Agent Behavior
* **No Yapping:** Be concise. Output the code, explain the changes in 1-2 sentences maximum, and move on. Do not apologize.
* **Destructive Actions:** If a user request requires deleting more than 50 lines of code, or dropping a database table, you MUST stop and ask for explicit confirmation before executing.
* **Verification:** After writing or modifying code, you MUST autonomously run the relevant build command (e.g., `nx build agent` or `nx typecheck web`) to verify your changes before telling the user you are done.


## Architecture Decision Records (ADRs)

We document all major architectural, tooling, and structural decisions as ADRs in the `/docs/adr/` directory.

* **Reading Context:** If you are asked to implement a new architectural pattern, introduce a new library, or fundamentally alter the data flow between the React frontend and Go services, you MUST first search and read the relevant files in `/docs/adr/` to ensure your approach does not violate past decisions.
* **Creating New ADRs:** If the user asks you to write or draft a new ADR, you MUST use the exact format defined in `/docs/adr/0000-example.md`. 
* **Numbering:** When creating a new ADR, look at the existing files in `/docs/adr/` and increment the ID number by 1 (e.g., if the latest is `0004-xxx.md`, name the new one `0005-xxx.md`).


## Git Workflow
* **Commits:** We use Conventional Commits (e.g., `feat(agent): add websocket parser`, `fix(front): resolve CORS issue`). Keep the body of the commit concise.
* **Branches:** Never commit directly to `main`. If asked to create a feature, create a branch named `feat/your-feature-name` or `fix/issue-description`.
