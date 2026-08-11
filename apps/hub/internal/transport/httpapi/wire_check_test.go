package httpapi

import (
	"encoding/json"
	"testing"
	"time"

	api "github.com/harmost/api/gen"
	"github.com/harmost/hub/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Guards the wire format against the hand-written DTOs these generated types
// replaced (ADR 0010). The migration was meant to be invisible to the front.
func TestWireFormatUnchanged(t *testing.T) {
	at := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	exit := int32(0)

	t.Run("agent zero value omits optionals", func(t *testing.T) {
		b, _ := json.Marshal(toAgentResponse(domain.Agent{}))
		assert.JSONEq(t, `{"id":"","name":"","description":"","version":"",
			"hostname":"","status":"","created_at":"0001-01-01T00:00:00Z"}`, string(b))
	})

	t.Run("agent populated", func(t *testing.T) {
		cpu, mem := float32(12.5), int64(2048)
		rc := int32(3)
		b, _ := json.Marshal(toAgentResponse(domain.Agent{
			Model:  domain.Model{ID: "a1", CreatedAt: at},
			Name:   "n", Status: domain.AgentStatusOnline, LastSeenAt: &at,
			CpuUsagePercent: &cpu, MemoryUsedBytes: &mem, RunningContainers: &rc,
		}))
		assert.JSONEq(t, `{"id":"a1","name":"n","description":"","version":"",
			"hostname":"","status":"online","last_seen_at":"2026-08-08T12:00:00Z",
			"created_at":"2026-08-08T12:00:00Z","cpu_usage_percent":12.5,
			"memory_used_bytes":2048,"running_containers":3}`, string(b))
	})

	t.Run("job with zero exit code keeps the field", func(t *testing.T) {
		b, _ := json.Marshal(toJobResponse(domain.Job{
			Model: domain.Model{ID: "j1", CreatedAt: at},
			AgentID: "a1", State: domain.JobStateSucceeded,
			Spec:     domain.JobSpec{Image: "alpine", Command: []string{"echo"}},
			ExitCode: &exit, StartedAt: &at,
		}))
		assert.JSONEq(t, `{"id":"j1","agent_id":"a1","state":"succeeded",
			"spec":{"image":"alpine","command":["echo"]},"message":"","exit_code":0,
			"started_at":"2026-08-08T12:00:00Z","created_at":"2026-08-08T12:00:00Z"}`, string(b))
	})

	t.Run("job spec omits every unset optional", func(t *testing.T) {
		b, _ := json.Marshal(specToAPI(domain.JobSpec{Image: "alpine"}))
		assert.JSONEq(t, `{"image":"alpine"}`, string(b))
	})

	t.Run("job spec round-trips through the domain mapper", func(t *testing.T) {
		orig := domain.JobSpec{
			Image: "alpine", Command: []string{"sh", "-c"}, Args: []string{"ls"},
			Env:            map[string]string{"K": "V"},
			VolumeMounts:   []domain.VolumeMount{{HostPath: "/h", ContainerPath: "/c", ReadOnly: true}},
			ResourceLimits: &domain.ResourceLimits{MemoryBytes: 1 << 20, CPUCores: 1.5},
			WorkingDir:     "/w", NetworkMode: "host", Labels: map[string]string{"L": "1"},
			PullPolicy: domain.PullPolicyAlways, Privileged: true, TimeoutSeconds: 30,
		}
		assert.Equal(t, orig, specFromAPI(specToAPI(orig)))
	})

	t.Run("job log", func(t *testing.T) {
		b, _ := json.Marshal(toJobLogResponse(domain.JobLog{
			ID: 7, Line: "hi", Stream: domain.LogStreamStderr, Sequence: 2, Timestamp: at,
		}))
		assert.JSONEq(t, `{"id":7,"line":"hi","stream":"stderr","sequence":2,
			"timestamp":"2026-08-08T12:00:00Z"}`, string(b))
	})

	t.Run("agent token never leaks the hash", func(t *testing.T) {
		b, _ := json.Marshal(api.AgentToken{ID: "t1", Name: "n", CreatedAt: at})
		assert.JSONEq(t, `{"id":"t1","name":"n","created_at":"2026-08-08T12:00:00Z"}`, string(b))
		assert.NotContains(t, string(b), "hash")
	})

	t.Run("error envelope", func(t *testing.T) {
		b, _ := json.Marshal(api.Error{Error: "nope"})
		assert.JSONEq(t, `{"error":"nope"}`, string(b))
	})
}
