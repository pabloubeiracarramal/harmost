package docker

// Manager tracks running jobs: Dispatch starts them, Cancel stops them.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SendFunc delivers an AgentMessage up the hub stream. It must be safe for
// concurrent use and must not block.
type SendFunc func(*harmostv1.AgentMessage)

type Manager struct {
	docker *Docker

	mu   sync.Mutex
	jobs map[string]context.CancelFunc
}

func NewManager(d *Docker) *Manager {
	return &Manager{docker: d, jobs: make(map[string]context.CancelFunc)}
}

// Dispatch runs a job in the background. A dispatch for a job that is still
// running is ignored — the hub may re-send after a stream reconnect.
// ctx bounds the job's lifetime (daemon shutdown), on top of which the spec's
// TimeoutSeconds is enforced.
func (m *Manager) Dispatch(ctx context.Context, jobID string, spec *harmostv1.JobSpec, send SendFunc) {
	jobCtx, cancel := context.WithCancel(ctx)
	if spec.TimeoutSeconds > 0 {
		jobCtx, cancel = context.WithTimeout(ctx, time.Duration(spec.TimeoutSeconds)*time.Second)
	}

	m.mu.Lock()
	if _, exists := m.jobs[jobID]; exists {
		m.mu.Unlock()
		cancel()
		return
	}
	m.jobs[jobID] = cancel
	m.mu.Unlock()

	go m.run(jobCtx, cancel, jobID, spec, send)
}

// Cancel stops a running job. Returns false if the job is not running.
func (m *Manager) Cancel(jobID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cancel, ok := m.jobs[jobID]; ok {
		cancel()
		return true
	}
	return false
}

func (m *Manager) run(ctx context.Context, cancel context.CancelFunc, jobID string, spec *harmostv1.JobSpec, send SendFunc) {

	defer func() {
		m.mu.Lock()
		delete(m.jobs, jobID)
		m.mu.Unlock()
		cancel()
	}()

	status := func(state harmostv1.JobState, message string, exitCode int32) {
		send(&harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_StatusUpdate{
			StatusUpdate: &harmostv1.JobStatusUpdate{
				JobId:     jobID,
				State:     state,
				Timestamp: timestamppb.Now(),
				Message:   message,
				ExitCode:  exitCode,
			},
		}})
	}

	var seq atomic.Int64
	logLine := func(stream harmostv1.LogStream, line string) {
		send(&harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_LogChunk{
			LogChunk: &harmostv1.LogChunk{
				JobId:     jobID,
				Line:      line,
				Timestamp: timestamppb.Now(),
				Stream:    stream,
				Sequence:  seq.Add(1),
			},
		}})
	}

	status(harmostv1.JobState_JOB_STATE_ACCEPTED, "", 0)

	exitCode, err := m.docker.Run(ctx, jobID, spec,
		func(state harmostv1.JobState, message string) { status(state, message, 0) },
		logLine,
	)

	switch {
	case err == nil && exitCode == 0:
		status(harmostv1.JobState_JOB_STATE_SUCCEEDED, "", 0)
	case err == nil:
		status(harmostv1.JobState_JOB_STATE_FAILED, fmt.Sprintf("exited with code %d", exitCode), int32(exitCode))
	case errors.Is(err, context.DeadlineExceeded):
		status(harmostv1.JobState_JOB_STATE_TIMED_OUT, fmt.Sprintf("timed out after %ds", spec.TimeoutSeconds), 0)
	case errors.Is(err, context.Canceled):
		status(harmostv1.JobState_JOB_STATE_CANCELLED, "", 0)
	default:
		status(harmostv1.JobState_JOB_STATE_FAILED, err.Error(), 0)
	}

}

// RunningJobs reports how many jobs are currently executing.
func (m *Manager) RunningJobs() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.jobs)
}

// RunningJobIDs returns the IDs of jobs currently executing, for the
// AgentHello reconciliation set.
func (m *Manager) RunningJobIDs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]string, 0, len(m.jobs))
	for id := range m.jobs {
		ids = append(ids, id)
	}
	return ids
}
