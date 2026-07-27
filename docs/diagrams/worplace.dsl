workspace "Harmost" "CI/CD orchestration SaaS — C4 model derived from CodeGraph analysis" {

    model {
        developer = person "Developer / DevOps Engineer" "Installs agents on their own machines, dispatches Docker-based jobs, and monitors live status and logs from the browser."

        github = softwareSystem "GitHub" "OAuth2 identity provider and profile source (github.com + api.github.com). Verified in code: httpapi/auth.go." {
            tags "External"
        }
        gitProvider = softwareSystem "Git Provider Webhooks" "Assumed External System — webhook-triggered dispatch (GitHub/GitLab push/PR) is documented as planned in docs/architecture.md; no implementing source exists in the graph." {
            tags "External" "Planned"
        }
        dockerEngine = softwareSystem "Docker Engine" "Container runtime on the user's machine. The agent drives it via the moby SDK (pull, create, start, wait, logs, stop, remove)." {
            tags "External"
        }

        harmost = softwareSystem "Harmost" "CI/CD orchestration SaaS: a cloud hub dispatches Docker-based jobs to user-hosted agents and streams live status and logs back to the browser." {

            front = container "Front" "Web UI: agent dashboard, job dispatch form, job list, live log viewer. Auth via JWT obtained from the hub's GitHub OAuth flow." "React 19, TypeScript, Vite, TanStack Router" {
                tags "Browser"
            }

            hub = container "Hub" "Cloud backend: REST + WebSocket server for the UI, gRPC bidi-stream server , org-scoped event fan-out." "Go (chi, grpc-go, GORM)" {
                httpApi = component "HTTP API" "httpapi.Server — REST routes under /api/v1/*, GitHub OAuth + device-flow endpoints, JWT bearer middleware." "Go, chi"
                wsHandler = component "WebSocket Handler" "handleWebSocket — upgrades /ws?token = , validateent's org on the event bus and pushes events." "Go"
                grpcApi = component "gRPC API" "grpcapi.Server — AgentService.Connect bidi-stream handler: hello/registration, status updates, batched log ingestion, heartbeats; implements httpapi.Dispatcher." "Go, grpc-go"
                streamRegistry = component "Agent Stream Registry" "In-memory, mutex-guarded map of agentID → thread-safe send function for the live gRPC stream. Basis of Dispatch/Cancel/Connected." "Go"
                eventBus = component "Event Bus" "events.Bus — in-process pub/sub keyed by org ID; non-blocking best-effort delivery (drops events on full 64-slot subscriber buffers)." "Go"
                jobService = component "JobService" "Job dispatch, SQL-guarded state transitions, reconcileweeping for offline agents." "Go"
                agentService = component "AgentService" "Agent registration/upsert on hello, online/offline lifecycle, heartbeat metrics snapshots." "Go"
                userService = component "UserService" "GitHub sign-up/login upsert; transactional first-login creation of user + personal org + owner membership." "Go"
                jobLogService = component "JobLogService" "Batched ingestion and retrieval of job log chunks." "Go"
                authJwt = component "Auth (JWT)" "Signing and validation of user+org claims (24h expiry); shared by REST middleware and WS handler." "Go"
                repositories = component "Repositories" "GORM data access for users, orgs, memberships, agents, agent tokens, jobs, job logs, device codes." "Go, GORM"
            }

            agentDaemon = container "Agent" "System-service daemon on the user's machine." "Go (kardianos/service, moby SDK)" {
                cli = component "CLI" "Cobra root command: 'pair', 'run', 'containers'. Entry point for both interactive use and service installation." "Go, cobra"
                pairing = component "Pairing" "runPair — OAuth2 device flow against the hub over REST: authorize, prompt user with verification URL, poll /device/token until approved, persist the agent token." "Go"
                configStore = component "Config Store" "Load/Save of hub address, gRPC address, and agent token as JSON in the OS user config dir (harmost/config.json, 0600)." "Go"
                daemonProgram = component "Daemon Program" "daemon.AgentProgram — service.Interface lifecycle; owns the outer loop: load config, init Docker (degrading gracefully if unavailable), reconnect to the hub with capped exponential backoff." "Go, kardianos/service"
                grpcClient = component "gRPC Client" "grpc.Client — opens the bidi stream with bearer-token metadata, sends AgentHello with running-job IDs; Send() queues outbound messages so terminal statuses survive a reconnect; handleMessage routes DispatchJob/CancelJob/Ping." "Go, grpc-go"
                jobManager = component "Job Manager" "docker.Manager — tracks running jobs (jobID → cancel func), dedupes re-dispatch after reconnect, enforces per-job timeouts, emits status updates and sequenced log lines via SendFunc." "Go"
                dockerRunner = component "Docker Runner" "docker.Docker + Run — one job's container lifecycle: pull (per pull policy) → create → start → follow logs (stdout/stderr demultiplexed line-by-line) → wait → force-remove; stop-with-deadline on cancel/timeout." "Go, moby SDK"
                metricsCollector = component "Metrics Collector" "metrics.Collect — CPU/memory/disk snapshots via gopsutil, sent to the hub as heartbeat payloads over the stream." "Go, gopsutil"
            }

            # ── Component relationships (agent) ──────────────────────────────────────
            developer -> cli "Runs 'agent pair <hub-url>', installs/starts the service"
            cli -> pairing "pair subcommand"
            cli -> daemonProgram "run/service subcommand"
            pairing -> hub "POST /api/v1/device/authorize, then polls /api/v1/device/token" "REST"
            pairing -> configStore "Persists hub addr + agent token on approval"
            daemonProgram -> configStore "Loads config at startup"
            daemonProgram -> grpcClient "Connect(target, token) in reconnect loop"
            daemonProgram -> dockerRunner "Initializes Docker SDK client, pings daemon"
            grpcClient -> hub "Bidi stream: hello, status, logs, heartbeats / dispatch, cancel, ping" "gRPC"
            grpcClient -> jobManager "Dispatch / Cancel from HubMessages; passes Send as the emit callback"
            grpcClient -> metricsCollector "Collects snapshot per heartbeat"
            jobManager -> dockerRunner "Run(ctx, jobID, spec) to completion"
            dockerRunner -> dockerEngine "Image pull, container create/start/wait/stop/remove, log follow" "Docker API"

            postgres = container "PostgreSQL" "Users, orgs, memberships, agent registry + tokens, jobs, job logs, device codes. Schema via goose migrations; accessed through GORM." "PostgreSQL" {
                tags "Database"
            }
        }

        # ── Context / container relationships ────────────────────────────────
        developer -> front "Monitors agents, dispatches and cancels jobs, watches live logs" "HTTPS"
        developer -> agentDaemon "Installs as a system service; runs 'agent pair' (device flow)"
        front -> hub "Commands and queries" "REST /api/v1/* (JWT bearer)"
        hub -> front "Agent/job/log events" "WebSocket /ws?token = <jwt>"
        agentDaemon -> hub "Initiates and holds bidi stream: AgentHello, status updates, log chunks, heartbeats; receives dispatch/cancel" "gRPC AgentService.Connect (bearer agent token)"
        hub -> github "Redirects for OAuth2 login; exchanges code; fetches user profile" "HTTPS"
        # ── Component relationships (hub) ─────────────────────────────────────
        front -> httpApi "Job/agent CRUD, dispatch, cancel, log history" "REST (JWT bearer)"
        front -> wsHandler "Subscribes to live events" "WebSocket"
        agentDaemon -> grpcApi "Bidi stream" "gRPC"
        httpApi -> authJwt "Validates bearer tokens; issues JWT after OAuth callback"
        wsHandler -> authJwt "Validates ?token = query parameter"
        httpApi -> github "OAuth2 code exchange + profile fetch" "HTTPS"
        httpApi -> userService "SignUpOrLogin, GetByID"
        httpApi -> agentService "List, GetByID"
        httpApi -> jobService "Dispatch, ListByOrg, GetByID, HandleStatusUpdate (dispatch-failure fallback)"
        httpApi -> jobLogService "ListByJob (log history)"
        httpApi -> grpcApi "Dispatch / Cancel / Connected — via the Dispatcher interface (no package import)"
        wsHandler -> eventBus "Subscribe(orgID)"
        grpcApi -> streamRegistry "add/remove/get live agent send functions"
        grpcApi -> agentService "Connect / UpdateOnConnect / Disconnect / HandleHeartbeat"
        grpcApi -> jobService "HandleStatusUpdate, ReconcileAgent"
        grpcApi -> jobLogService "IngestChunks (batched flush)"
        grpcApi -> eventBus "Publish agent.connected/disconnected/heartbeat, job.log"
        jobService -> eventBus "Publish job.status"
        jobService -> repositories "Job persistence with guarded state updates"
        agentService -> repositories "Agent registry persistence"
        userService -> repositories "User/org persistence (transactional)"
        jobLogService -> repositories "Log chunk persistence"
        repositories -> postgres "SQL" "GORM"
    }

    views {
        systemContext harmost "L1-SystemContext" {
            include *
            autolayout lr
        }
        container harmost "L2-Containers" {
            include *
            autolayout lr
        }
        component hub "L3-HubComponents" {
            include *
            autolayout lr
        }
        component agentDaemon "L3-AgentComponents" {
            include *
            autolayout lr
        }
        styles {
            element "Person" {
                shape person
                background #08427b
                color #ffffff
            }
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
            }
            element "Database" {
                shape cylinder
            }
            element "Browser" {
                shape webbrowser
            }
            element "External" {
                background #999999
                color #ffffff
            }
            element "Planned" {
                border dashed
                background #bbbbbb
            }
        }
    }
}