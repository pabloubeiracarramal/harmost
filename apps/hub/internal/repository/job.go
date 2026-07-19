package repository

import (
	"context"
	"time"

	"github.com/harmost/hub/internal/domain"
	"gorm.io/gorm"
)

type JobRepo struct {
	db *gorm.DB
}

func (r *JobRepo) Create(ctx context.Context, job *domain.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *JobRepo) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	var j domain.Job
	err := r.db.WithContext(ctx).First(&j, "id = ?", id).Error
	return &j, notFound(err)
}

func (r *JobRepo) ListByOrg(ctx context.Context, orgID string) ([]domain.Job, error) {
	var jobs []domain.Job
	err := r.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepo) ListByAgent(ctx context.Context, agentID string) ([]domain.Job, error) {
	var jobs []domain.Job
	err := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Find(&jobs).Error
	return jobs, err
}

// UpdateState applies a state transition unless the job is already terminal.
// The SQL guard is the authority under concurrent writers (stream handler,
// reconcile, sweeper); the returned bool reports whether the update applied.
func (r *JobRepo) UpdateState(ctx context.Context, id string, state domain.JobState, msg string, exitCode *int32, finishedAt *time.Time) (bool, error) {
	updates := map[string]any{
		"state":   state,
		"message": msg,
	}
	if exitCode != nil {
		updates["exit_code"] = *exitCode
	}
	if finishedAt != nil {
		updates["finished_at"] = finishedAt
	}
	res := r.db.WithContext(ctx).Model(&domain.Job{}).
		Where("id = ? AND state NOT IN ?", id, domain.TerminalJobStates).
		Updates(updates)
	return res.RowsAffected > 0, res.Error
}

func (r *JobRepo) ListActiveByAgent(ctx context.Context, agentID string, createdBefore time.Time) ([]domain.Job, error) {
	var jobs []domain.Job
	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND state NOT IN ? AND created_at < ?", agentID, domain.TerminalJobStates, createdBefore).
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepo) ListActiveForOfflineAgents(ctx context.Context, seenBefore time.Time) ([]domain.Job, error) {
	var jobs []domain.Job
	err := r.db.WithContext(ctx).
		Joins("JOIN agents ON agents.id = jobs.agent_id").
		Where("jobs.state NOT IN ? AND agents.status = ? AND agents.last_seen_at < ?",
			domain.TerminalJobStates, domain.AgentStatusOffline, seenBefore).
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepo) SetStarted(ctx context.Context, id string, at time.Time) (bool, error) {
	res := r.db.WithContext(ctx).Model(&domain.Job{}).
		Where("id = ? AND state NOT IN ?", id, domain.TerminalJobStates).
		Updates(map[string]any{
			"state":      domain.JobStateRunning,
			"started_at": at,
		})
	return res.RowsAffected > 0, res.Error
}
