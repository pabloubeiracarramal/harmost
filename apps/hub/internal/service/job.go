package service

import (
	"context"
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

// ListByOrg returns all jobs for an org ordered by created_at DESC.
func (s *JobService) ListByOrg(ctx context.Context, orgID string) ([]domain.Job, error) {
	return s.jobRepo.ListByOrg(ctx, orgID)
}

// GetByID returns a single job by ID.
func (s *JobService) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	return s.jobRepo.GetByID(ctx, id)
}
