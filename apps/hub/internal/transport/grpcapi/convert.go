package grpcapi

import (
	"time"

	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/harmost/hub/internal/domain"
)

// ─── proto → domain ──────────────────────────────────────────────────────────

func protoStateToJobState(s harmostv1.JobState) domain.JobState {
	switch s {
	case harmostv1.JobState_JOB_STATE_ACCEPTED:
		return domain.JobStateAccepted
	case harmostv1.JobState_JOB_STATE_PULLING_IMAGE:
		return domain.JobStatePullingImage
	case harmostv1.JobState_JOB_STATE_CREATING_CONTAINER:
		return domain.JobStateCreatingContainer
	case harmostv1.JobState_JOB_STATE_STARTING_CONTAINER:
		return domain.JobStateStartingContainer
	case harmostv1.JobState_JOB_STATE_RUNNING:
		return domain.JobStateRunning
	case harmostv1.JobState_JOB_STATE_STOPPING:
		return domain.JobStateStopping
	case harmostv1.JobState_JOB_STATE_CANCELLED:
		return domain.JobStateCancelled
	case harmostv1.JobState_JOB_STATE_SUCCEEDED:
		return domain.JobStateSucceeded
	case harmostv1.JobState_JOB_STATE_FAILED:
		return domain.JobStateFailed
	case harmostv1.JobState_JOB_STATE_TIMED_OUT:
		return domain.JobStateTimedOut
	default:
		return domain.JobStateUnspecified
	}
}

func protoChunkToJobLog(c *harmostv1.LogChunk) domain.JobLog {
	ts := time.Now()
	if c.Timestamp != nil {
		ts = c.Timestamp.AsTime()
	}
	return domain.JobLog{
		JobID:     c.JobId,
		Line:      c.Line,
		Stream:    protoStreamToLogStream(c.Stream),
		Sequence:  c.Sequence,
		Timestamp: ts,
	}
}

func protoStreamToLogStream(s harmostv1.LogStream) domain.LogStream {
	switch s {
	case harmostv1.LogStream_LOG_STREAM_STDOUT:
		return domain.LogStreamStdout
	case harmostv1.LogStream_LOG_STREAM_STDERR:
		return domain.LogStreamStderr
	default:
		return domain.LogStreamUnspecified
	}
}

// ─── domain → proto ──────────────────────────────────────────────────────────

func jobToProto(j *domain.Job) *harmostv1.DispatchJobRequest {
	return &harmostv1.DispatchJobRequest{
		JobId: j.ID,
		Spec:  jobSpecToProto(j.Spec),
	}
}

func jobSpecToProto(s domain.JobSpec) *harmostv1.JobSpec {
	spec := &harmostv1.JobSpec{
		Image:          s.Image,
		Command:        s.Command,
		Args:           s.Args,
		Env:            s.Env,
		Labels:         s.Labels,
		WorkingDir:     s.WorkingDir,
		NetworkMode:    s.NetworkMode,
		Privileged:     s.Privileged,
		TimeoutSeconds: s.TimeoutSeconds,
		PullPolicy:     domainPullPolicyToProto(s.PullPolicy),
	}
	if s.ResourceLimits != nil {
		spec.ResourceLimits = &harmostv1.ResourceLimits{
			MemoryBytes: s.ResourceLimits.MemoryBytes,
			CpuCores:    s.ResourceLimits.CPUCores,
		}
	}
	for _, vm := range s.VolumeMounts {
		spec.VolumeMounts = append(spec.VolumeMounts, &harmostv1.VolumeMount{
			HostPath:      vm.HostPath,
			ContainerPath: vm.ContainerPath,
			ReadOnly:      vm.ReadOnly,
		})
	}
	return spec
}

func domainPullPolicyToProto(p domain.PullPolicy) harmostv1.PullPolicy {
	switch p {
	case domain.PullPolicyAlways:
		return harmostv1.PullPolicy_PULL_POLICY_ALWAYS
	case domain.PullPolicyIfNotPresent:
		return harmostv1.PullPolicy_PULL_POLICY_IF_NOT_PRESENT
	case domain.PullPolicyNever:
		return harmostv1.PullPolicy_PULL_POLICY_NEVER
	default:
		return harmostv1.PullPolicy_PULL_POLICY_UNSPECIFIED
	}
}

