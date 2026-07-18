package domain

import (
	"context"
	"slices"
	"time"
)

type JobState string

const (
	JobStateUnspecified       JobState = "unspecified"
	JobStateAccepted          JobState = "accepted"
	JobStatePullingImage      JobState = "pulling_image"
	JobStateCreatingContainer JobState = "creating_container"
	JobStateStartingContainer JobState = "starting_container"
	JobStateRunning           JobState = "running"
	JobStateStopping          JobState = "stopping"
	JobStateCancelled         JobState = "cancelled"
	JobStateSucceeded         JobState = "succeeded"
	JobStateFailed            JobState = "failed"
	JobStateTimedOut          JobState = "timed_out"
)

// TerminalJobStates are the states a job can never leave. The repository's
// guarded updates and IsTerminal must agree on this set.
var TerminalJobStates = []JobState{
	JobStateCancelled,
	JobStateSucceeded,
	JobStateFailed,
	JobStateTimedOut,
}

func IsTerminal(s JobState) bool {
	return slices.Contains(TerminalJobStates, s)
}

type PullPolicy string

const (
	PullPolicyUnspecified  PullPolicy = "unspecified"
	PullPolicyAlways       PullPolicy = "always"
	PullPolicyIfNotPresent PullPolicy = "if_not_present"
	PullPolicyNever        PullPolicy = "never"
)

type VolumeMount struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only"`
}

type ResourceLimits struct {
	MemoryBytes int64   `json:"memory_bytes"`
	CPUCores    float32 `json:"cpu_cores"`
}

type JobSpec struct {
	Image          string            `json:"image"`
	Command        []string          `json:"command,omitempty"`
	Args           []string          `json:"args,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	VolumeMounts   []VolumeMount     `json:"volume_mounts,omitempty"`
	ResourceLimits *ResourceLimits   `json:"resource_limits,omitempty"`
	WorkingDir     string            `json:"working_dir,omitempty"`
	NetworkMode    string            `json:"network_mode,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	PullPolicy     PullPolicy        `json:"pull_policy,omitempty"`
	Privileged     bool              `json:"privileged,omitempty"`
	TimeoutSeconds int32             `json:"timeout_seconds,omitempty"`
}

type Job struct {
	Model
	OrgID      string   `gorm:"type:uuid;not null;index"`
	AgentID    string   `gorm:"type:uuid;not null;index"`
	State      JobState `gorm:"type:text;not null;default:'accepted'"`
	Spec       JobSpec  `gorm:"serializer:json"`
	Message    string
	ExitCode   *int32
	StartedAt  *time.Time
	FinishedAt *time.Time

	Org   Org   `gorm:"foreignKey:OrgID"`
	Agent Agent `gorm:"foreignKey:AgentID"`
}

type JobStatusInput struct {
	JobID     string
	State     JobState
	Message   string
	ExitCode  *int32
	Timestamp time.Time
}

type JobRepository interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id string) (*Job, error)
	ListByOrg(ctx context.Context, orgID string) ([]Job, error)
	ListByAgent(ctx context.Context, agentID string) ([]Job, error)
	UpdateState(ctx context.Context, id string, state JobState, msg string, exitCode *int32, finishedAt *time.Time) (bool, error)
	SetStarted(ctx context.Context, id string, at time.Time) (bool, error)
}

type JobService interface {
	Dispatch(ctx context.Context, orgID, agentID string, spec JobSpec) (*Job, error)
	HandleStatusUpdate(ctx context.Context, in JobStatusInput) error
	ListByOrg(ctx context.Context, orgID string) ([]Job, error)
	GetByID(ctx context.Context, id string) (*Job, error)
}
