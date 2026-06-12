package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/harmost/hub/internal/domain"
)

type agentResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	Hostname    string     `json:"hostname"`
	Status      string     `json:"status"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func toAgentResponse(a domain.Agent) agentResponse {
	return agentResponse{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Version:     a.Version,
		Hostname:    a.Hostname,
		Status:      string(a.Status),
		LastSeenAt:  a.LastSeenAt,
		CreatedAt:   a.CreatedAt,
	}
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	agents, err := s.svc.Agent.List(r.Context(), orgID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list agents")
		return
	}
	out := make([]agentResponse, len(agents))
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
