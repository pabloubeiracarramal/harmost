package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	api "github.com/harmost/api/gen"
	"github.com/harmost/hub/internal/domain"
)

func toAgentResponse(a domain.Agent) api.Agent {
	return api.Agent{
		ID:                a.ID,
		Name:              a.Name,
		Description:       a.Description,
		Version:           a.Version,
		Hostname:          a.Hostname,
		Status:            api.AgentStatus(a.Status),
		LastSeenAt:        a.LastSeenAt,
		CreatedAt:         a.CreatedAt,
		CPUUsagePercent:   a.CpuUsagePercent,
		MemoryUsedBytes:   a.MemoryUsedBytes,
		MemoryTotalBytes:  a.MemoryTotalBytes,
		DiskUsedBytes:     a.DiskUsedBytes,
		DiskTotalBytes:    a.DiskTotalBytes,
		RunningContainers: a.RunningContainers,
	}
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	agents, err := s.svc.Agent.List(r.Context(), orgID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list agents")
		return
	}
	out := make([]api.Agent, len(agents))
	for i, a := range agents {
		out[i] = toAgentResponse(a)
	}
	jsonOK(w, out)
}

func (s *Server) getAgent(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	agent, err := s.svc.Agent.GetByID(r.Context(), orgID, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "agent not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to get agent")
		return
	}
	jsonOK(w, toAgentResponse(*agent))
}

func (s *Server) deleteAgent(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := s.svc.Agent.Delete(r.Context(), orgID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "agent not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to delete agent")
		return
	}

	// Best-effort: the agent is already gone from the caller's perspective
	// either way.
	_ = s.svc.AgentToken.RevokeByAgentID(r.Context(), orgID, id)
	s.dispatcher.Kick(id)

	w.WriteHeader(http.StatusNoContent)
}
