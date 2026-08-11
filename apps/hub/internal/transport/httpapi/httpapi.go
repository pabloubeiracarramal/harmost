package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/harmost/hub/internal/auth"
	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/events"
	"github.com/harmost/hub/internal/platform"
	"github.com/harmost/hub/internal/service"
)

// Dispatcher is satisfied by grpcapi.Server — keeps httpapi free of a grpcapi import.
type Dispatcher interface {
	Dispatch(ctx context.Context, agentID string, job *domain.Job) error
	Cancel(ctx context.Context, agentID, jobID string) error
	Connected(agentID string) bool
	WatchContainers(ctx context.Context, agentID string) error
	UnwatchContainers(ctx context.Context, agentID string) error
}

type Server struct {
	svc         *service.Services
	dispatcher  Dispatcher
	bus         *events.Bus
	cfg         platform.Config
	publicLimit *ipRateLimiter
}

func New(svc *service.Services, d Dispatcher, bus *events.Bus, cfg platform.Config) *Server {
	return &Server{
		svc:        svc,
		dispatcher: d,
		bus:        bus,
		cfg:        cfg,
		// 0.5 tokens/sec refill, burst 10 — device-flow polling (>=5s interval)
		// stays comfortably under this; login/authorize bursts absorb the spike.
		publicLimit: newIPRateLimiter(0.5, 10),
	}
}

// Routes returns the fully configured chi router.
func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public — OAuth + device flow (no auth required), rate-limited per IP.
	r.Group(func(r chi.Router) {
		r.Use(s.publicLimit.middleware)

		r.Get("/auth/github", s.handleGitHubLogin)
		r.Get("/auth/github/callback", s.handleGitHubCallback)
		r.Post("/api/v1/device/authorize", s.handleDeviceAuthorize)
		r.Post("/api/v1/device/token", s.handleDeviceToken)
	})

	// WebSocket — auth via ?token= query param.
	r.Get("/ws", s.handleWebSocket)

	// Authenticated routes.
	r.Group(func(r chi.Router) {
		r.Use(s.jwtMiddleware)

		r.Post("/api/v1/device/approve", s.handleDeviceApprove)

		r.Get("/api/v1/me", s.getMe)

		r.Get("/api/v1/agents", s.listAgents)
		r.Get("/api/v1/agents/{id}", s.getAgent)
		r.Post("/api/v1/agents/{id}/containers/watch", s.watchAgentContainers)
		r.Post("/api/v1/agents/{id}/containers/unwatch", s.unwatchAgentContainers)

		r.Get("/api/v1/agent-tokens", s.listAgentTokens)
		r.Post("/api/v1/agent-tokens/{id}/revoke", s.revokeAgentToken)

		r.Post("/api/v1/jobs", s.dispatchJob)
		r.Get("/api/v1/jobs", s.listJobs)
		r.Get("/api/v1/jobs/{id}", s.getJob)
		r.Post("/api/v1/jobs/{id}/cancel", s.cancelJob)
		r.Get("/api/v1/jobs/{id}/logs", s.getJobLogs)
	})

	return r
}

// ─── context keys ─────────────────────────────────────────────────────────────

type ctxKey string

const (
	claimsKey ctxKey = "claims"
	orgIDKey  ctxKey = "orgID"
)

func claimsFromCtx(ctx context.Context) *auth.Claims {
	c, _ := ctx.Value(claimsKey).(*auth.Claims)
	return c
}

func orgIDFromCtx(ctx context.Context) string {
	c := claimsFromCtx(ctx)
	if c == nil {
		return ""
	}
	return c.OrgID
}

// ─── middleware ───────────────────────────────────────────────────────────────

func (s *Server) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == "" {
			jsonError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		claims, err := auth.Validate(tokenStr, s.cfg.JWTSecret)
		if err != nil {
			jsonError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ─── response helpers ────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
