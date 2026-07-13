package docker

// One job's container lifecycle: pull → create → start → stream logs → wait.

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// StatusFunc receives intermediate lifecycle states (pulling, creating, …).
// Terminal states (succeeded/failed/cancelled/timed out) are the caller's
// responsibility — only it knows whether a context cancellation was a user
// cancel or a timeout.
type StatusFunc func(state harmostv1.JobState, message string)

// LogFunc receives one log line at a time, already demultiplexed by stream.
type LogFunc func(stream harmostv1.LogStream, line string)

// Run executes a job's container to completion and returns its exit code.
// A non-nil error means the lifecycle itself failed (pull error, ctx
// cancelled, …); a normal container exit with non-zero code is (code, nil).
// The container is force-removed before returning, whatever the outcome.
func (d *Docker) Run(ctx context.Context, jobID string, spec *harmostv1.JobSpec, report StatusFunc, logLine LogFunc) (int, error) {

	if err := d.pull(ctx, spec, report); err != nil {
		return 0, err
	}

	report(harmostv1.JobState_JOB_STATE_CREATING_CONTAINER, "")
	cfg, hostCfg := toContainerConfig(jobID, spec)
	created, err := d.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     cfg,
		HostConfig: hostCfg,
	})

	if err != nil {
		return 0, fmt.Errorf("create container: %w", err)
	}

	defer func() {
		// The job ctx may already be cancelled — removal gets its own deadline.
		rmCtx, rmCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer rmCancel()
		if _, err := d.cli.ContainerRemove(rmCtx, created.ID, client.ContainerRemoveOptions{Force: true}); err != nil {
			report(harmostv1.JobState_JOB_STATE_UNSPECIFIED, fmt.Sprintf("remove container: %v", err))
		}
	}()

	// Register the wait before starting so the exit can't be missed.
	waitRes := d.cli.ContainerWait(ctx, created.ID, client.ContainerWaitOptions{
		Condition: container.WaitConditionNextExit,
	})

	report(harmostv1.JobState_JOB_STATE_STARTING_CONTAINER, "")
	if _, err := d.cli.ContainerStart(ctx, created.ID, client.ContainerStartOptions{}); err != nil {
		return 0, fmt.Errorf("start container: %w", err)
	}

	report(harmostv1.JobState_JOB_STATE_RUNNING, "")
	logsDone := d.streamLogs(ctx, created.ID, logLine)

	select {
	case res := <-waitRes.Result:
		<-logsDone // drain remaining output before reporting the exit
		if res.Error != nil {
			return 0, fmt.Errorf("container wait: %s", res.Error.Message)
		}
		return int(res.StatusCode), nil
	case err := <-waitRes.Error:
		return 0, fmt.Errorf("container wait: %w", err)
	case <-ctx.Done():
		report(harmostv1.JobState_JOB_STATE_STOPPING, "")
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer stopCancel()
		if _, err := d.cli.ContainerStop(stopCtx, created.ID, client.ContainerStopOptions{}); err != nil {
			report(harmostv1.JobState_JOB_STATE_UNSPECIFIED, fmt.Sprintf("stop container: %v", err))
		}
		return 0, ctx.Err()

	}

}

// pull fetches spec.Image according to spec.PullPolicy (default: if-not-present).
func (d *Docker) pull(ctx context.Context, spec *harmostv1.JobSpec, report StatusFunc) error {
	switch spec.PullPolicy {
	case harmostv1.PullPolicy_PULL_POLICY_NEVER:
		return nil
	case harmostv1.PullPolicy_PULL_POLICY_IF_NOT_PRESENT, harmostv1.PullPolicy_PULL_POLICY_UNSPECIFIED:
		_, err := d.cli.ImageInspect(ctx, spec.Image)
		if err == nil {
			return nil
		}
		if !cerrdefs.IsNotFound(err) {
			return fmt.Errorf("inspect image: %w", err)
		}
	}

	report(harmostv1.JobState_JOB_STATE_PULLING_IMAGE, spec.Image)
	resp, err := d.cli.ImagePull(ctx, spec.Image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	if err := resp.Wait(ctx); err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	return nil
}

// streamLogs follows the container's output, demultiplexes stdout/stderr and
// emits one LogFunc call per line. The returned channel closes once the log
// stream has been fully drained (i.e. the container exited).
func (d *Docker) streamLogs(ctx context.Context, containerID string, logLine LogFunc) <-chan struct{} {
	done := make(chan struct{})

	logs, err := d.cli.ContainerLogs(ctx, containerID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		logLine(harmostv1.LogStream_LOG_STREAM_STDERR, fmt.Sprintf("agent: cannot attach logs: %v", err))
		close(done)
		return done
	}

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()

	go scanLines(outR, harmostv1.LogStream_LOG_STREAM_STDOUT, logLine)
	go scanLines(errR, harmostv1.LogStream_LOG_STREAM_STDERR, logLine)

	go func() {
		defer close(done)
		defer logs.Close()
		_, err := stdcopy.StdCopy(outW, errW, logs)
		outW.CloseWithError(err)
		errW.CloseWithError(err)
	}()
	return done
}

func scanLines(r io.Reader, stream harmostv1.LogStream, logLine LogFunc) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		logLine(stream, sc.Text())
	}
	// A cancelled job closes the log stream mid-read — that's not worth a
	// synthetic error line, but a genuine scanner failure (e.g. a line over
	// the buffer cap) truncates output and must be surfaced.
	if err := sc.Err(); err != nil && !errors.Is(err, context.Canceled) {
		logLine(stream, fmt.Sprintf("agent: log stream error: %v", err))
	}
}
