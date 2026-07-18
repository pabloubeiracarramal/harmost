# M2 — Job lifecycle & live events (issues #10, #25)

## Context

M1 made jobs execute in Docker, but the lifecycle is fragile and invisible: jobs stay stuck forever if their agent dies, dispatching to an offline agent creates a job row and then fails it with a 502, a job spanning an agent reconnect silently loses all its status/log messages (they're written to the previous connection's abandoned `sendCh`), and job progress never reaches the frontend because the event bus only knows `agent.*` events. M2 fixes lifecycle resilience (#10) and fans job status/logs out over the WebSocket (#25).

**Delivery:** one branch `feat/job-lifecycle-events` off `develop`, single PR into `develop`, conventional commits per work item below.

## Design decisions

1. **Fail-on-disconnect vs survive-reconnect → reconcile-on-reconnect + grace sweeper.** No immediate sweep on disconnect (would kill jobs that survive a brief blip). Instead:
   - `AgentHello` gains `repeated string running_job_ids = 5`. On (re)connect the hub schedules a reconcile ~15s later (`time.AfterFunc`, `context.Background()`): any non-terminal DB job for that agent **not** in the hello set **and created before the hello** → `failed` ("job lost: agent reconnected without it"). The `created_at < helloAt` cutoff is load-bearing: reconcile fires on *every* connect, and without it any job dispatched in the 15s window after connect (allowed — the agent passes the `Connected` check) would be failed while its container runs. The 15s delay lets terminal statuses queued agent-side during the outage land first; the terminal guard (decision 3) then makes reconcile a no-op for them.
   - Background sweeper in the hub (30s tick, started from `main.go`): non-terminal jobs whose agent is `offline` with `last_seen_at` older than 2 min → `failed` ("agent offline"). `SetOffline` also sets `last_seen_at = now()` so the grace clock starts at disconnect. **Hub startup marks all agents offline** (`MarkAllOffline`, preserving `last_seen_at`): after a hub crash the stream handler's deferred `Disconnect` never ran, so agents linger `online` in the DB and the sweeper's `status='offline'` filter would never match them — at boot the registry is empty, so offline is true by definition, and reconnecting agents flip back online immediately.
2. **Agent send path → process-lifetime channels owned by `Client`.** `statusCh` (cap 256) and `logCh` (cap 1024) created once in `grpc.New`, not per-`Connect`. `(*Client).Send(msg)` (satisfies `docker.SendFunc`) routes StatusUpdate/Pong → statusCh, LogChunk → logCh; non-blocking, drop+log on full — except a **terminal** StatusUpdate on a full statusCh evicts the oldest queued message instead of being dropped (a dropped terminal status would make reconcile mislabel a finished job as "lost"). `Connect`'s loop drains statusCh with priority. On `stream.Send` failure of a StatusUpdate, stash it in `c.pendingStatus` and resend first on next connection. Manager's captured `send` closure is now valid for the process lifetime — logs and progress statuses are best-effort during outages; terminal statuses survive anything short of a >256-message pile-up of terminals, and reconcile is the backstop.
3. **Idempotency → SQL `WHERE state NOT IN (terminal)` guard in the repo**, returning `applied bool` from `RowsAffected` — authoritative under concurrency between stream handler, reconcile, and sweeper. Terminal set moves to `domain` (`domain.IsTerminal`); delete `isTerminal` from `service/job.go:50`.
4. **Events → extend `events.Event` with `JobID string` + `Payload any`** (both `omitempty`). `job.status` published from **inside `JobService`** so all write paths (gRPC update, dispatch-failure fallback, reconcile, sweeper) emit uniformly and only when the update applied. `job.log` published from the grpcapi `flush()` closure, **one event per job per flush batch** (500ms/500-line batching already exists → ≤2 events/s/job; per-line would flood the 64-slot subscriber buffer).
5. **Per-job WS log filtering → deferred to M3** (jobs UI). Events carry `job_id`; client-side filtering suffices at MVP scale.

## Work items (ordered commits on one branch)

### 1. `feat(hub): idempotent job status handling and live job events` (#25)

- `apps/hub/internal/domain/job.go` — add `TerminalJobStates` + `IsTerminal(JobState) bool`; `JobRepository` interface: `UpdateState(...) (bool, error)`, `SetStarted(...) (bool, error)`.
- `apps/hub/internal/repository/job.go` — add `AND state NOT IN ?` to both updates; return `RowsAffected > 0`.
- `apps/hub/internal/events/bus.go` — add `JobStatus = "job.status"`, `JobLog = "job.log"`; extend `Event` (bus.go:16) with `JobID`, `Payload`; add payload structs:
  ```go
  type JobStatusPayload struct { State string; Message string `json:",omitempty"`; ExitCode *int32 `json:",omitempty"` }
  type LogLine struct { Line, Stream string; Sequence int64; Timestamp time.Time }
  type JobLogPayload struct { Lines []LogLine }
  ```
- `apps/hub/internal/service/job.go` — `JobService` gains a narrow `Publisher interface { Publish(events.Event) }`. `HandleStatusUpdate` (job.go:26): `GetByID` first (need OrgID/AgentID for the event; early-return if already terminal), guarded update, publish `job.status` iff applied.
- `apps/hub/internal/service/service.go` — `service.New` accepts the bus (`main.go` already constructs the bus first, main.go:29).
- `apps/hub/internal/transport/grpcapi/connect.go` — `flush()` (connect.go:105): after successful `IngestChunks`, group `logBuf` by JobID, publish one `job.log` per job (orgID/agentID in scope).
- `apps/hub/internal/transport/httpapi/ws.go` — no functional change; delete dead `wsEvent` (ws.go:95).
- No migration (no schema change).

WS client JSON shape:
```json
{"type":"job.status","agent_id":"…","job_id":"…","at":"…","payload":{"state":"failed","message":"exited with code 2","exit_code":2}}
{"type":"job.log","agent_id":"…","job_id":"…","at":"…","payload":{"lines":[{"line":"hello","stream":"stdout","sequence":1,"timestamp":"…"}]}}
```

Note: `exit_code` is absent for successful jobs — proto3 `int32` can't distinguish 0 from unset, and `connect.go:149` only maps non-zero values. Frontend treats `state=succeeded` as exit 0. (Making the proto field `optional` is deferred; not worth a wire change for M2.)

### 2. `feat(proto): running_job_ids in AgentHello`

- `libs/harmost-proto/proto/harmost/v1/agent.proto` — `AgentHello` field `repeated string running_job_ids = 5;`. Regenerate via buf (Nx target), commit `gen/`. Backward compatible.

### 3. `feat(agent): survive reconnects with process-lifetime send queue`

- `apps/agent/internal/docker/manager.go` — add `RunningJobIDs() []string` (keys of `m.jobs` under mu).
- `apps/agent/internal/grpc/client.go` — the core refactor (client.go:62-69 removed):
  - `Client` gains `statusCh`, `logCh` (made in `New`), `pendingStatus *harmostv1.AgentMessage`.
  - `(*Client).Send` routes by payload type as per decision 2; passed to `mgr.Dispatch` and `handleMessage`.
  - `Connect`: hello includes `RunningJobIds` (nil `mgr` → empty); after hello, resend `pendingStatus` if set; main loop drains statusCh before logCh (nested select); on send failure of a StatusUpdate, stash to `pendingStatus`.
- `apps/agent/CLAUDE.md` — update the sendCh gotcha (buffer semantics changed).

### 4. `feat(hub): offline dispatch rejection, reconcile-on-reconnect, orphan sweeper` (#10)

- `grpcapi/grpcapi.go` — `(s *Server) Connected(agentID string) bool` (registry lookup); add to the `Dispatcher` interface in `httpapi/httpapi.go:19`.
- `httpapi/jobs.go` `dispatchJob` (jobs.go:66) — before creating the job: agent lookup scoped to caller's org → 404 (also closes today's cross-org dispatch hole); `!Connected` → **409** `{"error":"agent is not connected"}`, no row created. Keep the create-then-fail fallback for the check-to-send race (now emits a `job.status` failed event via item 1) — and set `Timestamp: time.Now()` in that fallback's `JobStatusInput`: today it's zero-valued, and item 1's `finishedAt = &in.Timestamp` would persist `0001-01-01`.
- `domain/job.go` + `repository/job.go` — new repo methods:
  `ListActiveByAgent(ctx, agentID, createdBefore time.Time)` (state NOT IN terminal `AND created_at < ?` — the reconcile cutoff from decision 1) and `ListActiveForOfflineAgents(ctx, seenBefore time.Time)` (JOIN agents; `status='offline' AND last_seen_at < ?`).
- `repository/agent.go` — `SetOffline` (agent.go:46) also sets `last_seen_at = now()`; new `MarkAllOffline(ctx)` (`UPDATE agents SET status='offline' WHERE status='online'`, `last_seen_at` untouched so the grace clock reflects real last contact), called once at hub startup (decision 1 — recovers agents left `online` by a hub crash).
- `service/job.go` — `ReconcileAgent(ctx, agentID, runningJobIDs, helloAt)` and `SweepOrphans(ctx, grace)`; both use the guarded `UpdateState(failed)` + publish per applied row (shared `applyStatus` helper with `HandleStatusUpdate`).
- `grpcapi/connect.go` — after hello registration (~:58): capture `helloAt := time.Now()`, then `time.AfterFunc(15*time.Second, func(){ ReconcileAgent(context.Background(), agent.ID, ids, helloAt) })` (re-queries active jobs at fire time; safe if the stream dropped again, and jobs dispatched after the hello are excluded by the cutoff). Fires with `context.Background()`, so it may run during shutdown against a closing DB pool — harmless, at worst a log line.
- `cmd/hub/main.go` — call `MarkAllOffline` after DB init; sweeper goroutine (30s tick, 2m grace), cancelled at shutdown. ~12 lines inline.

### 5. `test(hub,agent): first unit tests` (testify — none exist yet in the repo)

- `apps/hub/internal/service/job_test.go` — fake `JobRepository` + fake `Publisher`: applied update publishes; duplicate/late terminal is dropped with no event; RUNNING → `SetStarted`; `ReconcileAgent` fails only jobs absent from the running set **and leaves jobs created after `helloAt` untouched**; `SweepOrphans`.
- `apps/hub/internal/events/bus_test.go` — org scoping, unsubscribe, drop-when-full.
- `apps/agent/internal/grpc/client_test.go` — `Send` routing + drop semantics (incl. terminal status evicting oldest on full statusCh).
- Repo SQL-guard tests deferred (no Postgres test harness); covered by E2E.

### 6. `docs: sync status.md` (+ roadmap M2 check-off) — same turn as implementation, per standing rule.

## Verification

Build/vet/test via Nx (`nx run hub:build`, `agent:build`, `hub:lint`, `agent:lint`, `hub:test`, `agent:test`).

Manual E2E (`nx run workspace:dev`, `nx run agent:serve`, `apps/hub/api.http`, `websocat "ws://localhost:8080/ws?token=<jwt>"`):

1. **Happy path:** dispatch a busybox echo job → WS shows `job.status` accepted→…→succeeded + batched `job.log` events.
2. **Offline rejection:** stop agent, dispatch → 409, no job row.
3. **Hub restart mid-job:** dispatch `sleep 60`, restart hub → agent reconnects with the job in `running_job_ids`, reconcile no-op, terminal status lands.
4. **Terminal status across outage:** dispatch `sleep 20`, hub down t≈10s–30s → queued `succeeded` delivered on reconnect, beats the 15s reconcile.
5. **Agent killed mid-job:** `kill -9` agent, restart quickly → hello lacks the job → `failed` ~15s after reconnect, with WS event.
6. **Sweeper:** kill agent, leave down → job `failed` ~2m after disconnect.
7. **Idempotency:** via grpcui send `JobStatusUpdate(running)` for a succeeded job → state unchanged, no WS event.
8. **Fresh-dispatch survives reconcile:** restart agent, dispatch a `sleep 30` within 15s of reconnect → job unaffected when reconcile fires, completes normally.
9. **Hub crash recovery:** `kill -9` hub while an agent is connected, keep agent down, restart hub → agent flipped offline at boot, its job swept `failed` once the 2m grace passes.

## Explicitly deferred

- Per-job WS log filtering → M3 (client filters on `job_id` meanwhile).
- Proto ack/resume for guaranteed log delivery across outages → post-MVP; statuses are protected, logs best-effort.
- Re-dispatch of `accepted` jobs on reconnect → not required by #10; pre-dispatch check shrinks the window, reconcile fails anything truly lost.
- Partial index on active jobs / DB-backed repo tests → when scale or a Postgres test harness justifies it.
- **Sweeper vs long partition:** if the sweeper fails a job at 2m but the agent was alive behind a network partition, the agent's later `succeeded` is dropped by the terminal guard and nobody sends `CancelJob` — the container runs to completion invisibly. Accepted for MVP; cancel-on-reconcile-mismatch is the post-MVP fix.
- `optional int32 exit_code` in proto so exit 0 is representable on the wire → post-MVP (see WS shape note in item 1).
