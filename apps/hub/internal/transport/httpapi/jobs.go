package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/harmost/hub/internal/domain"
)

// ─── request / response types ────────────────────────────────────────────────

type dispatchRequest struct {
	AgentID string         `json:"agent_id"`
	Spec    domain.JobSpec `json:"spec"`
}

type jobResponse struct {
	ID         string         `json:"id"`
	AgentID    string         `json:"agent_id"`
	State      string         `json:"state"`
	Spec       domain.JobSpec `json:"spec"`
	Message    string         `json:"message"`
	ExitCode   *int32         `json:"exit_code,omitempty"`
	StartedAt  *time.Time     `json:"started_at,omitempty"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type jobLogResponse struct {
	ID        uint64    `json:"id"`
	Line      string    `json:"line"`
	Stream    string    `json:"stream"`
	Sequence  int64     `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
}

func toJobResponse(j domain.Job) jobResponse {
	return jobResponse{
		ID:         j.ID,
		AgentID:    j.AgentID,
		State:      string(j.State),
		Spec:       j.Spec,
		Message:    j.Message,
		ExitCode:   j.ExitCode,
		StartedAt:  j.StartedAt,
		FinishedAt: j.FinishedAt,
		CreatedAt:  j.CreatedAt,
	}
}

func toJobLogResponse(l domain.JobLog) jobLogResponse {
	return jobLogResponse{
		ID:        l.ID,
		Line:      l.Line,
		Stream:    string(l.Stream),
		Sequence:  l.Sequence,
		Timestamp: l.Timestamp,
	}
}

// ─── handlers ────────────────────────────────────────────────────────────────

func (s *Server) dispatchJob(w http.ResponseWriter, r *http.Request) {
	orgID := orgIDFromCtx(r.Context())

	var req dispatchRequest
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

	job, err := s.svc.Job.Dispatch(r.Context(), orgID, req.AgentID, req.Spec)
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
	out := make([]jobResponse, len(jobs))
	for i, j := range jobs {
		out[i] = toJobResponse(j)
	}
	jsonOK(w, out)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := s.svc.Job.GetByID(r.Context(), id)
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

	job, err := s.svc.Job.GetByID(r.Context(), id)
	if err != nil || job.OrgID != orgID {
		if errors.Is(err, domain.ErrNotFound) || err == nil {
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
	id := chi.URLParam(r, "id")
	logs, err := s.svc.JobLog.ListByJob(r.Context(), id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to get logs")
		return
	}
	out := make([]jobLogResponse, len(logs))
	for i, l := range logs {
		out[i] = toJobLogResponse(l)
	}
	jsonOK(w, out)
}
