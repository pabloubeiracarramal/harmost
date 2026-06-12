package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/service"
)

// Dispatcher is satisfied by grpcapi.Server — keeps httpapi free of a grpcapi import.
type Dispatcher interface {
	Dispatch(ctx context.Context, agentID string, job *domain.Job) error
}

type Server struct {
	svc        *service.Services
	dispatcher Dispatcher
}

func New(svc *service.Services, d Dispatcher) *Server {
	return &Server{svc: svc, dispatcher: d}
}

// Routes returns the fully configured chi router.
func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(orgIDMiddleware)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/agents", s.listAgents)
		r.Get("/agents/{id}", s.getAgent)

		r.Post("/jobs", s.dispatchJob)
		r.Get("/jobs", s.listJobs)
		r.Get("/jobs/{id}", s.getJob)
		r.Get("/jobs/{id}/logs", s.getJobLogs)
	})

	return r
}

// ─── context helpers ─────────────────────────────────────────────────────────

type ctxKey string

const orgIDKey ctxKey = "orgID"

func orgIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("x-org-id")
		if id == "" {
			jsonError(w, http.StatusUnauthorized, "missing x-org-id header")
			return
		}
		ctx := context.WithValue(r.Context(), orgIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func orgIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(orgIDKey).(string)
	return v
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
