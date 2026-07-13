package repository

import (
	"context"

	"github.com/harmost/hub/internal/domain"
	"gorm.io/gorm"
)

type JobLogRepo struct {
	db *gorm.DB
}

// CreateBatch persists log chunks in batches of 500 to avoid large insert statements.
func (r *JobLogRepo) CreateBatch(ctx context.Context, logs []domain.JobLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(logs, 500).Error
}

func (r *JobLogRepo) ListByJob(ctx context.Context, jobID string) ([]domain.JobLog, error) {
	var logs []domain.JobLog
	err := r.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Order("sequence ASC").
		Find(&logs).Error
	return logs, err
}
