package service

import (
	"context"
	"errors"
	"time"

	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/events"
)

// Publisher is the narrow slice of the event bus JobService needs.
type Publisher interface {
	Publish(events.Event)
}

type JobService struct {
	jobRepo domain.JobRepository
	bus     Publisher
}

// Dispatch creates a new job in accepted state for the given agent.
func (s *JobService) Dispatch(ctx context.Context, orgID, agentID string, spec domain.JobSpec) (*domain.Job, error) {
	job := &domain.Job{
		OrgID:   orgID,
		AgentID: agentID,
		State:   domain.JobStateAccepted,
		Spec:    spec,
	}
	return job, s.jobRepo.Create(ctx, job)
}

// HandleStatusUpdate applies a state transition received from an agent.
func (s *JobService) HandleStatusUpdate(ctx context.Context, in domain.JobStatusInput) error {
	job, err := s.jobRepo.GetByID(ctx, in.JobID)
	if err != nil {
		return err
	}
	if domain.IsTerminal(job.State) {
		return nil // late or duplicate update for a finished job
	}
	return s.applyStatus(ctx, job, in)
}

// applyStatus performs the guarded state update and publishes job.status iff
// the update applied. job supplies OrgID/AgentID for the event; the SQL guard
// in the repository decides races, not the state read into job.
func (s *JobService) applyStatus(ctx context.Context, job *domain.Job, in domain.JobStatusInput) error {
	var applied bool
	var err error
	if in.State == domain.JobStateRunning {
		applied, err = s.jobRepo.SetStarted(ctx, in.JobID, in.Timestamp)
	} else {
		var finishedAt *time.Time
		if domain.IsTerminal(in.State) {
			t := in.Timestamp
			finishedAt = &t
		}
		applied, err = s.jobRepo.UpdateState(ctx, in.JobID, in.State, in.Message, in.ExitCode, finishedAt)
	}
	if err != nil || !applied {
		return err
	}

	s.bus.Publish(events.Event{
		Type:    events.JobStatus,
		OrgID:   job.OrgID,
		AgentID: job.AgentID,
		JobID:   job.ID,
		At:      time.Now(),
		Payload: events.JobStatusPayload{
			State:    string(in.State),
			Message:  in.Message,
			ExitCode: in.ExitCode,
		},
	})
	return nil
}

// ReconcileAgent fails jobs the DB believes are active on an agent but the
// agent did not report in its hello. Called ~15s after (re)connect so terminal
// statuses queued agent-side during an outage land first (the terminal guard
// then makes this a no-op for them). helloAt excludes jobs dispatched after
// the hello — they are alive on the new stream, not lost.
func (s *JobService) ReconcileAgent(ctx context.Context, agentID string, runningJobIDs []string, helloAt time.Time) error {
	jobs, err := s.jobRepo.ListActiveByAgent(ctx, agentID, helloAt)
	if err != nil {
		return err
	}
	running := make(map[string]struct{}, len(runningJobIDs))
	for _, id := range runningJobIDs {
		running[id] = struct{}{}
	}

	var errs []error
	for i := range jobs {
		if _, ok := running[jobs[i].ID]; ok {
			continue
		}
		errs = append(errs, s.applyStatus(ctx, &jobs[i], domain.JobStatusInput{
			JobID:     jobs[i].ID,
			State:     domain.JobStateFailed,
			Message:   "job lost: agent reconnected without it",
			Timestamp: time.Now(),
		}))
	}
	return errors.Join(errs...)
}

// SweepOrphans fails jobs whose agent has been offline longer than grace.
func (s *JobService) SweepOrphans(ctx context.Context, grace time.Duration) error {
	jobs, err := s.jobRepo.ListActiveForOfflineAgents(ctx, time.Now().Add(-grace))
	if err != nil {
		return err
	}
	var errs []error
	for i := range jobs {
		errs = append(errs, s.applyStatus(ctx, &jobs[i], domain.JobStatusInput{
			JobID:     jobs[i].ID,
			State:     domain.JobStateFailed,
			Message:   "agent offline",
			Timestamp: time.Now(),
		}))
	}
	return errors.Join(errs...)
}

// ListByOrg returns all jobs for an org ordered by created_at DESC.
func (s *JobService) ListByOrg(ctx context.Context, orgID string) ([]domain.Job, error) {
	return s.jobRepo.ListByOrg(ctx, orgID)
}

// GetByID returns a single job by ID, scoped to the caller's org. A job owned
// by another org is reported as not found rather than forbidden — callers must
// not be able to probe for job IDs outside their tenant.
//
// Agent-driven paths (status updates, reconciliation, the orphan sweeper) go
// through s.jobRepo directly; they are authenticated by agent token, not by a
// user's org claim.
func (s *JobService) GetByID(ctx context.Context, orgID, id string) (*domain.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if job.OrgID != orgID {
		return nil, domain.ErrNotFound
	}
	return job, nil
}
