package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	api "github.com/harmost/api/gen"
	"github.com/harmost/hub/internal/domain"
)

func toJobResponse(j domain.Job) api.Job {
	return api.Job{
		ID:         j.ID,
		AgentID:    j.AgentID,
		State:      api.JobState(j.State),
		Spec:       specToAPI(j.Spec),
		Message:    j.Message,
		ExitCode:   j.ExitCode,
		StartedAt:  j.StartedAt,
		FinishedAt: j.FinishedAt,
		CreatedAt:  j.CreatedAt,
	}
}

func toJobLogResponse(l domain.JobLog) api.JobLog {
	return api.JobLog{
		ID:        l.ID,
		Line:      l.Line,
		Stream:    api.LogStream(l.Stream),
		Sequence:  l.Sequence,
		Timestamp: l.Timestamp,
	}
}

// ─── handlers ────────────────────────────────────────────────────────────────

func (s *Server) dispatchJob(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())

	var req api.DispatchJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AgentID == "" {
		jsonError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	if req.Spec.Image == "" {
		jsonError(w, http.StatusBadRequest, "spec.image is required")
		return
	}
	spec := specFromAPI(req.Spec)

	// Reject before creating a row: unknown/foreign agent → 404 (lookup is
	// org-scoped), known but disconnected agent → 409.
	if _, err := s.svc.Agent.GetByID(r.Context(), orgID, req.AgentID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "agent not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to look up agent")
		return
	}
	if !s.dispatcher.Connected(req.AgentID) {
		jsonError(w, http.StatusConflict, "agent is not connected")
		return
	}

	job, err := s.svc.Job.Dispatch(r.Context(), orgID, req.AgentID, spec)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	// Fallback for the check-to-send race: the agent may drop between the
	// Connected check above and this send.
	if err := s.dispatcher.Dispatch(r.Context(), req.AgentID, job); err != nil {
		_ = s.svc.Job.HandleStatusUpdate(r.Context(), domain.JobStatusInput{
			JobID:     job.ID,
			State:     domain.JobStateFailed,
			Message:   "agent not connected: " + err.Error(),
			Timestamp: time.Now(),
		})
		jsonError(w, http.StatusBadGateway, "agent not connected")
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, toJobResponse(*job))
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	jobs, err := s.svc.Job.ListByOrg(r.Context(), orgID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list jobs")
		return
	}
	out := make([]api.Job, len(jobs))
	for i, j := range jobs {
		out[i] = toJobResponse(j)
	}
	jsonOK(w, out)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	job, err := s.svc.Job.GetByID(r.Context(), orgID, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "job not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to get job")
		return
	}
	jsonOK(w, toJobResponse(*job))
}

func (s *Server) cancelJob(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	job, err := s.svc.Job.GetByID(r.Context(), orgID, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "job not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to get job")
		return
	}
	if domain.IsTerminal(job.State) {
		jsonError(w, http.StatusConflict, "job already finished")
		return
	}
	if err := s.dispatcher.Cancel(r.Context(), job.AgentID, job.ID); err != nil {
		jsonError(w, http.StatusConflict, "agent is not connected")
		return
	}

	// The agent drives the state transition (stopping → cancelled) via the
	// gRPC stream; 202 reflects that cancellation is underway, not done.
	w.WriteHeader(http.StatusAccepted)
	jsonOK(w, toJobResponse(*job))
}

func (s *Server) getJobLogs(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	logs, err := s.svc.JobLog.ListByJob(r.Context(), orgID, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "job not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to get logs")
		return
	}
	out := make([]api.JobLog, len(logs))
	for i, l := range logs {
		out[i] = toJobLogResponse(l)
	}
	jsonOK(w, out)
}
