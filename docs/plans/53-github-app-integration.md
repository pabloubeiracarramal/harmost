# GitHub App integration — installation + repo listing (issue #53)

Sub-issue of #52 ("Define the Project concept"). Not yet implemented — this is a plan for review.

## Context

This is the first of three sub-issues under #52. The eventual goal is a "create Project from a GitHub repo" flow (#54, #55), deliberately built on a **GitHub App** rather than widening the existing login OAuth app's scope — chosen for future CI/CD headroom (installation-scoped repo grants, short-lived tokens, free webhooks/checks later for auto-deploy and PR statuses).

Confirmed today: the login OAuth flow (`apps/hub/internal/transport/httpapi/auth.go`) requests only `read:user user:email` and discards the GitHub access token right after fetching the profile — no repo-listing plumbing exists anywhere in the hub. This issue builds that from scratch, hub-only. No `Project` entity (#54) and no front-end (#55) — those consume this issue's `GET /api/v1/integrations/github/repos` endpoint later.

## Design decisions

1. **Auth context across the GitHub redirect.** The browser navigates away to `github.com` and back during installation, so no bearer JWT survives the round trip (same constraint the existing OAuth login already solves via a `state` cookie).
   - `GET /api/v1/integrations/github/install` (authenticated) returns JSON `{"install_url": "..."}` rather than redirecting directly, so the front can attach the `Authorization` header before navigating itself (`window.location = install_url`) — a plain 3xx here would lose the header. The URL embeds a short-lived (10 min) HMAC-signed `state` token carrying the caller's `org_id`.
   - `GET /api/v1/integrations/github/callback` (public, GitHub redirects the browser here) verifies `state`, recovers `org_id`, and upserts the installation.
   - `GET /api/v1/integrations/github/repos` (authenticated, org-scoped) mints a fresh installation token per request (no caching — not a hot path, simplest correct MVP choice) and paginates `GET /installation/repositories`.
2. **Callback payload.** GitHub's install callback only guarantees `installation_id`, `setup_action`, and `state` on the query string — not account login/type. So `HandleCallback` makes one extra App-JWT-authenticated `GET /app/installations/{id}` call to learn `account.login`/`account.type` before upserting, rather than trusting query params that don't exist.
3. **One installation per org for MVP** — unique constraint both ways (`org_id` and `installation_id`). No webhook-driven uninstall handling; a stale row if the App is uninstalled from GitHub's side is an accepted gap (webhooks are separately-tracked future scope, per #53's stated non-goals).
4. **Config stays optional.** New env vars (`GITHUB_APP_ID`, `GITHUB_APP_SLUG`, `GITHUB_APP_PRIVATE_KEY_FILE`) load via `getEnv` + empty defaults, not `requireEnv` like the login OAuth vars — the hub keeps booting in dev/CI without a registered App; only the new endpoints fail until configured. Matches the existing optional-feature precedent set by `GRPC_TLS_CERT_FILE`/`KEY_FILE`.
5. **Wire types are generated, never hand-written** (per `apps/hub/CLAUDE.md`) — `GitHubInstallResponse` and `GitHubRepo` go into `libs/harmost-api/openapi.yaml` first; handlers use the generated `api.*` types, with a small inline mapping from `domain.GitHubRepo` (three flat fields don't warrant a full `convert.go` treatment like `JobSpec` gets).

## Work items

### 1. Config — `apps/hub/internal/platform/platform.go`

Add three fields per decision 4:

```go
GitHubAppID             string // GITHUB_APP_ID
GitHubAppSlug           string // GITHUB_APP_SLUG — builds https://github.com/apps/<slug>/installations/new
GitHubAppPrivateKeyFile string // GITHUB_APP_PRIVATE_KEY_FILE — PEM path
```

### 2. Migration — `apps/hub/migrations/011_create_github_installations.sql`

Mirrors `004_create_agents.sql`'s style exactly (goose Up/Down, `UUID PRIMARY KEY DEFAULT gen_random_uuid()`, `REFERENCES orgs(id) ON DELETE CASCADE`, `TIMESTAMPTZ NOT NULL DEFAULT now()`):

```sql
-- +goose Up
CREATE TABLE github_installations (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    installation_id BIGINT      NOT NULL,
    account_login   TEXT        NOT NULL,
    account_type    TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_github_installations_org UNIQUE (org_id)
);
CREATE UNIQUE INDEX idx_github_installations_installation_id ON github_installations (installation_id);

-- +goose Down
DROP TABLE github_installations;
```

Upsert on `org_id` conflict (`ON CONFLICT (org_id) DO UPDATE ...`) handles both `install` and `update` `setup_action`s identically.

### 3. Domain — `apps/hub/internal/domain/github.go` (new)

Follows the `Org`/`Agent` pattern exactly. Naming uses `GitHub` (capital H) to match existing identifiers (`GitHubProfile`, `GitHubClientID`), not `Github`.

```go
type GitHubInstallation struct {
    Model
    OrgID          string `gorm:"type:uuid;not null;uniqueIndex"`
    InstallationID int64  `gorm:"not null;uniqueIndex"`
    AccountLogin   string `gorm:"not null"`
    AccountType    string `gorm:"not null"`
    Org Org `gorm:"foreignKey:OrgID"`
}

type GitHubInstallationRepository interface {
    Upsert(ctx context.Context, inst *GitHubInstallation) error
    GetByOrgID(ctx context.Context, orgID string) (*GitHubInstallation, error)
}

// GitHubRepo is deliberately a small DTO, not the full GitHub API payload.
type GitHubRepo struct {
    FullName      string
    Private       bool
    DefaultBranch string
}

type GitHubIntegrationService interface {
    InstallURL(ctx context.Context, orgID string) (string, error)
    HandleCallback(ctx context.Context, state string, installationID int64) error
    ListRepos(ctx context.Context, orgID string) ([]GitHubRepo, error)
}
```

(`HandleCallback` takes only `state`+`installationID` — per decision 2, it fetches account login/type itself via the App JWT rather than trusting query params that don't exist.)

### 4. Repository — `apps/hub/internal/repository/github.go` (new)

Thin GORM wrapper matching `repository/org.go`'s style, using `clause.OnConflict` for the upsert. Register `GitHubInstallation *GitHubInstallationRepo` on `Repos` in `repository.go`.

### 5. Service — `apps/hub/internal/service/github.go` (new)

Constructor-injected fields: `installRepo domain.GitHubInstallationRepository`, `appID`, `appSlug string`, `privateKey *rsa.PrivateKey`, `jwtSecret string`, `httpBaseURL string` (default `"https://api.github.com"`, overridable — the one deliberate deviation from `fetchGitHubProfile`'s hardcoded URL, needed so tests can point at an `httptest` server), `httpClient *http.Client` (default `http.DefaultClient`).

Four pieces:

- **State token** — new `SignInstallState(orgID, secret)` / `ValidateInstallState(token, secret)` in `apps/hub/internal/auth/jwt.go`, alongside (not merged into) the existing `Sign`/`Validate` — same `RegisteredClaims` + custom-field + HS256 pattern, 10 min TTL, distinct claims type (`InstallStateClaims{ RegisteredClaims; OrgID string }`) so it's not confused with session JWTs.
- **App JWT** — `iss=appID`, `iat=now-30s` (clock-skew buffer), `exp=now+9m` (under GitHub's 10 min cap), RS256 signed with `privateKey`.
- **Installation token exchange** — `POST {httpBaseURL}/app/installations/{id}/access_tokens` with `Authorization: Bearer <appJWT>`, same raw `net/http` + `io.ReadAll` + `json.Unmarshal` style as `fetchGitHubProfile`.
- **`HandleCallback`** — validates state → `GET {httpBaseURL}/app/installations/{id}` (App JWT auth) to read `account.login`/`account.type` → upserts.
- **`ListRepos`** — loads the org's installation (404 via `domain.ErrNotFound` if none) → mints an installation token → loops `GET {httpBaseURL}/installation/repositories?per_page=100&page=N` until a short page or a 500-repo/5-page cap is hit → maps to `[]domain.GitHubRepo`.

**Wiring change**: `service.New(db *gorm.DB, cfg platform.Config, bus *events.Bus)` — replaces the current `frontendURL string` param with the full `cfg` (superset; `DeviceFlowService.frontendURL` now reads `cfg.FrontendURL`). Parse the PEM once here (`os.ReadFile` + `jwt.ParseRSAPrivateKeyFromPEM`, `log.Fatalf` only if the path is set but unparseable — an unset path just leaves `privateKey` nil and GitHub-integration calls return a clear "not configured" error). Update the one call site in `cmd/hub/main.go`. Add `GitHubIntegration *GitHubIntegrationService` to `Services` and its compile-time interface check.

### 6. HTTP handlers — `apps/hub/internal/transport/httpapi/github.go` (new)

Per decision 5: add `GitHubInstallResponse {install_url}` and `GitHubRepo {full_name, private, default_branch}` schemas to `openapi.yaml` and use the generated `api.GitHubInstallResponse` / `[]api.GitHubRepo` types in the handlers (small inline mapping from `domain.GitHubRepo`).

```go
func (s *Server) handleGitHubInstall(w http.ResponseWriter, r *http.Request)  // authenticated, orgIDFromCtx
func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) // public, 307 to {FrontendURL}/settings?github=connected|error
func (s *Server) listGitHubRepos(w http.ResponseWriter, r *http.Request)      // authenticated, 404 via domain.ErrNotFound
```

Register in `httpapi.go`: `handleGitHubCallback` in the public rate-limited group (alongside `/auth/github/callback`); `handleGitHubInstall`/`listGitHubRepos` in the authenticated group (alongside `/api/v1/agents`).

### 7. OpenAPI — `libs/harmost-api/openapi.yaml`

New `integrations` tag; three paths (`/api/v1/integrations/github/install`, `/callback`, `/repos`) following the existing `/auth/github` redirect-doc template and the `/api/v1/agents` JSON template; two new schemas (`GitHubInstallResponse`, `GitHubRepo`). Then `nx run harmost-api:generate` and `nx run harmost-api:check` to confirm no drift.

### 8. Tests

- `apps/hub/internal/auth/jwt_test.go` — extend if it exists (check first) with `SignInstallState`/`ValidateInstallState` round-trip: valid, expired (construct claims directly, don't sleep), wrong secret.
- `apps/hub/internal/service/github_test.go` (new) — `fakeGitHubInstallationRepo` (mirrors `job_test.go`'s fake-repo style) + `httptest.NewServer` mocking `/app/installations/{id}/access_tokens`, `/app/installations/{id}`, and paginated `/installation/repositories`. Cover: `ListRepos` happy path + pagination + no-installation 404; `HandleCallback` upsert + bad/expired state rejected; App JWT has correct `iss`/`exp` bounds (throwaway `rsa.GenerateKey` in-test, no real App key needed).
- `apps/hub/internal/transport/httpapi/github_test.go` — if handler-level tests exist elsewhere (check `agents_test.go`), add cases for install (200 + `install_url`), callback (307 + correct redirect target), repos (200 / 404).

## Critical files

- `apps/hub/internal/domain/github.go` (new)
- `apps/hub/internal/repository/github.go` (new)
- `apps/hub/internal/service/github.go` (new)
- `apps/hub/internal/transport/httpapi/github.go` (new)
- `apps/hub/internal/auth/jwt.go` — add state-token sign/verify
- `apps/hub/internal/platform/platform.go` — new config fields
- `apps/hub/internal/service/service.go`, `apps/hub/cmd/hub/main.go` — `service.New` signature change
- `apps/hub/internal/transport/httpapi/httpapi.go` — route registration
- `apps/hub/internal/repository/repository.go` — register new repo
- `apps/hub/migrations/011_create_github_installations.sql` (new)
- `libs/harmost-api/openapi.yaml` — new paths + schemas

## Verification

- `nx run hub:lint`, `nx run hub:test`, `nx run hub:build`
- `nx run harmost-api:generate` then `nx run harmost-api:check` (must show no diff)
- Manual smoke test (requires registering a real GitHub App and setting the three new env vars — a one-time manual prerequisite, not part of this issue's code): apply migration 011, get a dev JWT via `cmd/devtoken`, `curl` `/api/v1/integrations/github/install`, follow `install_url` in a browser, confirm the redirect lands on `{FRONTEND_URL}/settings?github=connected`, then `curl` `/api/v1/integrations/github/repos` and confirm the installed repos come back.
