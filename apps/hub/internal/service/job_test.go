package service

import (
	"context"
	"testing"
	"time"

	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeJobRepo struct {
	job     *domain.Job  // returned by GetByID
	applied bool         // returned by UpdateState / SetStarted
	active  []domain.Job // returned by ListActiveByAgent
	orphans []domain.Job // returned by ListActiveForOfflineAgents

	setStartedCalls  int
	updatedIDs       []string
	gotState         domain.JobState
	gotMessage       string
	gotCreatedBefore time.Time
	gotSeenBefore    time.Time
}

func (f *fakeJobRepo) Create(ctx context.Context, job *domain.Job) error { return nil }
func (f *fakeJobRepo) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	return f.job, nil
}
func (f *fakeJobRepo) ListByOrg(ctx context.Context, orgID string) ([]domain.Job, error) {
	return nil, nil
}
func (f *fakeJobRepo) ListByAgent(ctx context.Context, agentID string) ([]domain.Job, error) {
	return nil, nil
}

func (f *fakeJobRepo) UpdateState(ctx context.Context, id string, state domain.JobState, msg string, exitCode *int32, finishedAt *time.Time) (bool, error) {
	f.updatedIDs = append(f.updatedIDs, id)
	f.gotState = state
	f.gotMessage = msg
	return f.applied, nil
}

func (f *fakeJobRepo) SetStarted(ctx context.Context, id string, at time.Time) (bool, error) {
	f.setStartedCalls++
	return f.applied, nil
}

func (f *fakeJobRepo) ListActiveByAgent(ctx context.Context, agentID string, createdBefore time.Time) ([]domain.Job, error) {
	f.gotCreatedBefore = createdBefore
	return f.active, nil
}

func (f *fakeJobRepo) ListActiveForOfflineAgents(ctx context.Context, seenBefore time.Time) ([]domain.Job, error) {
	f.gotSeenBefore = seenBefore
	return f.orphans, nil
}

type fakePublisher struct {
	events []events.Event
}

func (f *fakePublisher) Publish(e events.Event) { f.events = append(f.events, e) }

func testJob(id string, state domain.JobState) *domain.Job {
	j := &domain.Job{OrgID: "org-1", AgentID: "agent-1", State: state}
	j.ID = id
	return j
}

func TestHandleStatusUpdate_AppliedPublishes(t *testing.T) {
	repo := &fakeJobRepo{job: testJob("job-1", domain.JobStateRunning), applied: true}
	pub := &fakePublisher{}
	svc := &JobService{jobRepo: repo, bus: pub}

	err := svc.HandleStatusUpdate(context.Background(), domain.JobStatusInput{
		JobID: "job-1", State: domain.JobStateSucceeded, Timestamp: time.Now(),
	})

	require.NoError(t, err)
	require.Len(t, pub.events, 1)
	e := pub.events[0]
	assert.Equal(t, events.JobStatus, e.Type)
	assert.Equal(t, "org-1", e.OrgID)
	assert.Equal(t, "agent-1", e.AgentID)
	assert.Equal(t, "job-1", e.JobID)
	payload, ok := e.Payload.(events.JobStatusPayload)
	require.True(t, ok)
	assert.Equal(t, "succeeded", payload.State)
}

func TestHandleStatusUpdate_LateTerminalDropped(t *testing.T) {
	repo := &fakeJobRepo{job: testJob("job-1", domain.JobStateSucceeded), applied: true}
	pub := &fakePublisher{}
	svc := &JobService{jobRepo: repo, bus: pub}

	err := svc.HandleStatusUpdate(context.Background(), domain.JobStatusInput{
		JobID: "job-1", State: domain.JobStateRunning, Timestamp: time.Now(),
	})

	require.NoError(t, err)
	assert.Empty(t, repo.updatedIDs, "no update for an already-terminal job")
	assert.Zero(t, repo.setStartedCalls)
	assert.Empty(t, pub.events)
}

func TestHandleStatusUpdate_GuardedDuplicateNoEvent(t *testing.T) {
	// Job reads non-terminal but the SQL guard reports not-applied (a
	// concurrent writer won the race).
	repo := &fakeJobRepo{job: testJob("job-1", domain.JobStateRunning), applied: false}
	pub := &fakePublisher{}
	svc := &JobService{jobRepo: repo, bus: pub}

	err := svc.HandleStatusUpdate(context.Background(), domain.JobStatusInput{
		JobID: "job-1", State: domain.JobStateFailed, Timestamp: time.Now(),
	})

	require.NoError(t, err)
	assert.Empty(t, pub.events)
}

func TestHandleStatusUpdate_RunningUsesSetStarted(t *testing.T) {
	repo := &fakeJobRepo{job: testJob("job-1", domain.JobStateAccepted), applied: true}
	pub := &fakePublisher{}
	svc := &JobService{jobRepo: repo, bus: pub}

	err := svc.HandleStatusUpdate(context.Background(), domain.JobStatusInput{
		JobID: "job-1", State: domain.JobStateRunning, Timestamp: time.Now(),
	})

	require.NoError(t, err)
	assert.Equal(t, 1, repo.setStartedCalls)
	assert.Empty(t, repo.updatedIDs)
	require.Len(t, pub.events, 1)
}

func TestReconcileAgent_FailsOnlyJobsAbsentFromRunningSet(t *testing.T) {
	repo := &fakeJobRepo{
		active:  []domain.Job{*testJob("job-a", domain.JobStateRunning), *testJob("job-b", domain.JobStateRunning)},
		applied: true,
	}
	pub := &fakePublisher{}
	svc := &JobService{jobRepo: repo, bus: pub}

	helloAt := time.Now()
	err := svc.ReconcileAgent(context.Background(), "agent-1", []string{"job-a"}, helloAt)

	require.NoError(t, err)
	assert.Equal(t, []string{"job-b"}, repo.updatedIDs)
	assert.Equal(t, domain.JobStateFailed, repo.gotState)
	assert.Equal(t, "job lost: agent reconnected without it", repo.gotMessage)
	assert.Equal(t, helloAt, repo.gotCreatedBefore, "cutoff must be the hello time so later dispatches survive")
	require.Len(t, pub.events, 1)
	assert.Equal(t, "job-b", pub.events[0].JobID)
}

func TestSweepOrphans_FailsAndPublishes(t *testing.T) {
	repo := &fakeJobRepo{
		orphans: []domain.Job{*testJob("job-c", domain.JobStateRunning)},
		applied: true,
	}
	pub := &fakePublisher{}
	svc := &JobService{jobRepo: repo, bus: pub}

	grace := 2 * time.Minute
	err := svc.SweepOrphans(context.Background(), grace)

	require.NoError(t, err)
	assert.Equal(t, []string{"job-c"}, repo.updatedIDs)
	assert.Equal(t, "agent offline", repo.gotMessage)
	assert.WithinDuration(t, time.Now().Add(-grace), repo.gotSeenBefore, time.Second)
	require.Len(t, pub.events, 1)
	assert.Equal(t, "job-c", pub.events[0].JobID)
}
