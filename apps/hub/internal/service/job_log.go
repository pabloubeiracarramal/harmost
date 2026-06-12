package service

import (
	"context"

	"github.com/harmost/hub/internal/domain"
)

type JobLogService struct {
	jobLogRepo domain.JobLogRepository
}

// IngestChunks persists a batch of log lines. The transport layer is responsible
// for accumulating individual LogChunk messages before calling this.
func (s *JobLogService) IngestChunks(ctx context.Context, chunks []domain.JobLog) error {
	return s.jobLogRepo.CreateBatch(ctx, chunks)
}

// ListByJob returns all log lines for a job ordered by sequence.
func (s *JobLogService) ListByJob(ctx context.Context, jobID string) ([]domain.JobLog, error) {
	return s.jobLogRepo.ListByJob(ctx, jobID)
}
