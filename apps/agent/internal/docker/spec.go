package docker

import (
	"maps"
	"sort"

	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/moby/moby/api/types/container"
)

// JobIDLabel is stamped on every container this agent creates, so containers
// can be matched back to their job after an agent restart (orphan cleanup,
// metrics, the `containers` debug command).
const JobIDLabel = "harmost.job_id"

// toContainerConfig translates the hub's JobSpec into the two structs
// ContainerCreate expects.
//
// Field mapping follows the Kubernetes convention:
//   - Command → Entrypoint (overrides the image's ENTRYPOINT)
//   - Args    → Cmd        (overrides the image's CMD)
//   - Env map → "KEY=VALUE" slice, sorted so output is deterministic
//
// TimeoutSeconds is deliberately not mapped: timeouts are enforced by the
// job's context in the manager, not by Docker. The remaining post-MVP spec
// fields (volume mounts, resource limits, network mode, privileged) belong
// in the HostConfig and are not mapped yet.
func toContainerConfig(jobID string, spec *harmostv1.JobSpec) (*container.Config, *container.HostConfig) {
	env := make([]string, 0, len(spec.Env))
	for k, v := range spec.Env {
		env = append(env, k+"="+v)
	}
	sort.Strings(env)

	labels := make(map[string]string, len(spec.Labels)+1)
	maps.Copy(labels, spec.Labels)
	labels[JobIDLabel] = jobID

	cfg := &container.Config{
		Image:      spec.Image,
		Entrypoint: spec.Command,
		Cmd:        spec.Args,
		Env:        env,
		WorkingDir: spec.WorkingDir,
		Labels:     labels,
	}
	return cfg, &container.HostConfig{}
}
