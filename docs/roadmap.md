# MVP Roadmap — 8 weeks (solo dev)

> Written 2026-07-08. Tracked on GitHub: [milestones](https://github.com/pabloubeiracarramal/harmost/milestones) · [open issues](https://github.com/pabloubeiracarramal/harmost/issues).
> Working rule: anything not on the login → pair → dispatch → live logs path gets deferred.

## MVP definition

One demonstrable loop:

**Log in with GitHub → pair an agent → dispatch a container job from the UI → watch live logs stream → job finishes with exit code → cancel works.**

### Explicitly cut from MVP

- Multi-org switcher (schema is ready per ADR 0004; UI deferred)
- Additional OAuth providers (Google, GitLab)
- mTLS, `Pong` latency tracking, `SetHeartbeatInterval`
- Job spec extras: volumes, privileged, network modes, resource limits (image + command + env + timeout only)

## Where the codebase stands (2026-07-08)

Done: hub auth (GitHub OAuth → JWT, device flow, org-scoped agent tokens), gRPC
bidi stream server with status/log persistence, REST for agents + jobs, WS
fanout of agent events; agent CLI + daemon with reconnect and metric
heartbeats; front login/dashboard/agent detail with live status.

The gap: **jobs never execute.** The agent logs `DispatchJob` and does nothing
(`apps/agent/internal/grpc/client.go`). Job events never reach the WebSocket.
No jobs UI. No TLS on gRPC. No tests. No deployment story.

*(Update 2026-07-13: M1 closed the execution gap — jobs now run in Docker with
status/logs persisted. Remaining gaps are M2 onward.)*

## Plan

| # | Milestone | Due | Issues | Scope |
|---|-----------|-----|--------|-------|
| M1 ✅ | Agent Docker executor — **done 2026-07-13** | Jul 22 | [#16](https://github.com/pabloubeiracarramal/harmost/issues/16), [#17](https://github.com/pabloubeiracarramal/harmost/issues/17) | The product hinge. Docker SDK executor: pull → create → start → attach; every `JobState` transition; exit codes; `timeout_seconds`; `CancelJob`; log chunks with sequence numbers streamed up the existing gRPC stream. |
| M2 ✅ | Job lifecycle & live events — **done 2026-07-18** | Jul 29 | [#10](https://github.com/pabloubeiracarramal/harmost/issues/10), [#25](https://github.com/pabloubeiracarramal/harmost/issues/25) | Orphan sweeper (agent dies → non-terminal jobs `failed`), reject dispatch to offline agents, survive reconnect mid-job. Add `job.status` / `job.log` bus events and forward over the WebSocket. |
| M3 | Jobs UI | Aug 5 | [#21](https://github.com/pabloubeiracarramal/harmost/issues/21), [#22](https://github.com/pabloubeiracarramal/harmost/issues/22), [#26](https://github.com/pabloubeiracarramal/harmost/issues/26) | Dispatch form, jobs list, job detail with live log viewer (backfill + WS append). Session hardening: WS auto-reconnect, JWT expiry handling, `GET /me`. |
| M4 | Security & production hardening | Aug 12 | [#27](https://github.com/pabloubeiracarramal/harmost/issues/27), [#28](https://github.com/pabloubeiracarramal/harmost/issues/28) | TLS on gRPC (agent keeps `--insecure` for local dev), agent-token list/revoke UI, rate limiting on unauthenticated endpoints. |
| M5 | Deployment & packaging | Aug 19 | [#29](https://github.com/pabloubeiracarramal/harmost/issues/29), [#30](https://github.com/pabloubeiracarramal/harmost/issues/30), [#31](https://github.com/pabloubeiracarramal/harmost/issues/31) | Hub Dockerfile + hosted deploy (Fly.io/Railway) with managed Postgres and migrations on deploy; agent binaries via goreleaser + install script; CI (nx affected build/vet/test, buf checks). |
| M6 | Hardening, tests & beta | Sep 2 | [#32](https://github.com/pabloubeiracarramal/harmost/issues/32), [#33](https://github.com/pabloubeiracarramal/harmost/issues/33) | Tests where bugs are costly: token validation, JWT middleware, job state machine, executor integration test. Onboarding polish, README quickstart, 2–3 beta users. Week 8 is buffer. |

## Known blockers

None. (Docker Desktop WSL integration was enabled 2026-07-13; current state
lives in `docs/status.md`.)

## Why 2 months is realistic

M1 is the only genuinely hard build; every later milestone wires together
layers that already exist (log ingestion, event bus, WS server, REST). The
sequencing keeps the app demoable at the end of every milestone.
