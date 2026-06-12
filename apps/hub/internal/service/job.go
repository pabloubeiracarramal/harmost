package service

import (
	"context"
	"time"

	"github.com/harmost/hub/internal/domain"
)

type JobService struct {
	jobRepo domain.JobRepository
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
	var finishedAt *time.Time
	if isTerminal(in.State) {
		t := in.Timestamp
		finishedAt = &t
	}

	if in.State == domain.JobStateRunning {
		return s.jobRepo.SetStarted(ctx, in.JobID, in.Timestamp)
	}

	return s.jobRepo.UpdateState(ctx, in.JobID, in.State, in.Message, in.ExitCode, finishedAt)
}

// ListByOrg returns all jobs for an org ordered by created_at DESC.
func (s *JobService) ListByOrg(ctx context.Context, orgID string) ([]domain.Job, error) {
	return s.jobRepo.ListByOrg(ctx, orgID)
}

// GetByID returns a single job by ID.
func (s *JobService) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	return s.jobRepo.GetByID(ctx, id)
}

func isTerminal(s domain.JobState) bool {
	switch s {
	case domain.JobStateCancelled, domain.JobStateSucceeded, domain.JobStateFailed, domain.JobStateTimedOut:
		return true
	}
	return false
}
