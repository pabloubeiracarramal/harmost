package service

import (
	"context"
	"testing"

	"github.com/harmost/hub/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeJobLogRepo struct {
	logs []domain.JobLog
}

func (f *fakeJobLogRepo) CreateBatch(ctx context.Context, logs []domain.JobLog) error { return nil }
func (f *fakeJobLogRepo) ListByJob(ctx context.Context, jobID string) ([]domain.JobLog, error) {
	return f.logs, nil
}

func TestListByJob_ReturnsLogsForOwningOrg(t *testing.T) {
	jobRepo := &fakeJobRepo{job: testJob("job-1", domain.JobStateRunning)} // OrgID "org-1"
	logRepo := &fakeJobLogRepo{logs: []domain.JobLog{{JobID: "job-1", Line: "hello"}}}
	svc := &JobLogService{jobLogRepo: logRepo, jobRepo: jobRepo}

	logs, err := svc.ListByJob(context.Background(), "org-1", "job-1")

	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "hello", logs[0].Line)
}

func TestListByJob_ForeignOrgIsNotFound(t *testing.T) {
	jobRepo := &fakeJobRepo{job: testJob("job-1", domain.JobStateRunning)} // OrgID "org-1"
	logRepo := &fakeJobLogRepo{logs: []domain.JobLog{{JobID: "job-1", Line: "secret"}}}
	svc := &JobLogService{jobLogRepo: logRepo, jobRepo: jobRepo}

	logs, err := svc.ListByJob(context.Background(), "org-2", "job-1")

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, logs)
}
