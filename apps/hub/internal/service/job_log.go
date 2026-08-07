package service

import (
	"context"

	"github.com/harmost/hub/internal/domain"
)

type JobLogService struct {
	jobLogRepo domain.JobLogRepository
	jobRepo    domain.JobRepository
}

// IngestChunks persists a batch of log lines. The transport layer is responsible
// for accumulating individual LogChunk messages before calling this.
func (s *JobLogService) IngestChunks(ctx context.Context, chunks []domain.JobLog) error {
	return s.jobLogRepo.CreateBatch(ctx, chunks)
}

// ListByJob returns all log lines for a job ordered by sequence, scoped to the
// caller's org. A job owned by another org is reported as not found rather than
// forbidden — callers must not be able to probe for job IDs outside their tenant.
func (s *JobLogService) ListByJob(ctx context.Context, orgID, jobID string) ([]domain.JobLog, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job.OrgID != orgID {
		return nil, domain.ErrNotFound
	}
	return s.jobLogRepo.ListByJob(ctx, jobID)
}
