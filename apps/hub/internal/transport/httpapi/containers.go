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

func (s *Server) startContainer(w http.ResponseWriter, r *http.Request) {
	s.containerAction(w, r, "start")
}

func (s *Server) stopContainer(w http.ResponseWriter, r *http.Request) {
	s.containerAction(w, r, "stop")
}

func (s *Server) restartContainer(w http.ResponseWriter, r *http.Request) {
	s.containerAction(w, r, "restart")
}

func (s *Server) removeContainer(w http.ResponseWriter, r *http.Request) {
	s.containerAction(w, r, "remove")
}

// containerAction dispatches a lifecycle action to any container on the
// host — not scoped to harmost-dispatched jobs. Fire-and-forget: this
// response only means "sent to the agent"; the real outcome arrives over
// WS as agent.container_action, same split as job dispatch/status.
func (s *Server) containerAction(w http.ResponseWriter, r *http.Request, action string) {
	orgID := orgIDFromCtx(r.Context())
	agentID := chi.URLParam(r, "id")
	containerID := chi.URLParam(r, "containerId")

	if _, err := s.svc.Agent.GetByID(r.Context(), orgID, agentID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "agent not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to look up agent")
		return
	}
	if !s.dispatcher.Connected(agentID) {
		jsonError(w, http.StatusConflict, "agent is not connected")
		return
	}

	if err := s.dispatcher.ContainerAction(r.Context(), agentID, containerID, action); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to dispatch container action")
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
