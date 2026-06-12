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

func (r *JobRepo) UpdateState(ctx context.Context, id string, state domain.JobState, msg string, exitCode *int32, finishedAt *time.Time) error {
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
	return r.db.WithContext(ctx).Model(&domain.Job{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *JobRepo) SetStarted(ctx context.Context, id string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.Job{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"state":      domain.JobStateRunning,
			"started_at": at,
		}).Error
}
