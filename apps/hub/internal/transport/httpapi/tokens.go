package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	api "github.com/harmost/api/gen"
	"github.com/harmost/hub/internal/domain"
)

func (s *Server) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	tokens, err := s.svc.AgentToken.List(r.Context(), orgID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list tokens")
		return
	}
	// api.AgentToken deliberately has no TokenHash — even the hash never
	// leaves the hub.
	out := make([]api.AgentToken, len(tokens))
	for i, t := range tokens {
		out[i] = api.AgentToken{
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
