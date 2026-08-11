package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/harmost/hub/internal/domain"
)

func (s *Server) watchAgentContainers(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if _, err := s.svc.Agent.GetByID(r.Context(), orgID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "agent not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to look up agent")
		return
	}

	if err := s.dispatcher.WatchContainers(r.Context(), id); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to start watching containers")
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) unwatchAgentContainers(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if _, err := s.svc.Agent.GetByID(r.Context(), orgID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "agent not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to look up agent")
		return
	}

	if err := s.dispatcher.UnwatchContainers(r.Context(), id); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to stop watching containers")
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
