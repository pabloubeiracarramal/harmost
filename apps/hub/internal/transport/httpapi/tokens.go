package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/harmost/hub/internal/domain"
)

// agentTokenResponse deliberately omits TokenHash — even the hash never
// leaves the hub.
type agentTokenResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	AgentID    *string    `json:"agent_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

func (s *Server) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	tokens, err := s.svc.AgentToken.List(r.Context(), orgID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list tokens")
		return
	}
	out := make([]agentTokenResponse, len(tokens))
	for i, t := range tokens {
		out[i] = agentTokenResponse{
			ID:         t.ID,
			Name:       t.Name,
			AgentID:    t.AgentID,
			CreatedAt:  t.CreatedAt,
			LastUsedAt: t.LastUsedAt,
		}
	}
	jsonOK(w, out)
}

func (s *Server) revokeAgentToken(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	if err := s.svc.AgentToken.Revoke(r.Context(), orgID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "token not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to revoke token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
