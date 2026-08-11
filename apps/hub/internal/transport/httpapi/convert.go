package httpapi

import (
	api "github.com/harmost/api/gen"
	"github.com/harmost/hub/internal/domain"
)

// JobSpec is the one wire type that is also a persistence type: it is stored on
// domain.Job via `gorm:"serializer:json"`. Letting domain import the generated
// api package would invert the transport → service → repository layering, so the
// two representations stay separate and meet here.
//
// The structs are field-for-field identical (the spec marks JobSpec's optional
// scalars x-go-type-skip-optional-pointer precisely so they are), but Go cannot
// convert between them directly because []api.VolumeMount and
// []domain.VolumeMount are distinct slice types.

func specFromAPI(s api.JobSpec) domain.JobSpec {
	return domain.JobSpec{
		Image:          s.Image,
		Command:        s.Command,
		Args:           s.Args,
		Env:            s.Env,
		VolumeMounts:   volumeMountsFromAPI(s.VolumeMounts),
		ResourceLimits: resourceLimitsFromAPI(s.ResourceLimits),
		WorkingDir:     s.WorkingDir,
		NetworkMode:    s.NetworkMode,
		Labels:         s.Labels,
		PullPolicy:     domain.PullPolicy(s.PullPolicy),
		Privileged:     s.Privileged,
		TimeoutSeconds: s.TimeoutSeconds,
	}
}

func specToAPI(s domain.JobSpec) api.JobSpec {
	return api.JobSpec{
		Image:          s.Image,
		Command:        s.Command,
		Args:           s.Args,
		Env:            s.Env,
		VolumeMounts:   volumeMountsToAPI(s.VolumeMounts),
		ResourceLimits: resourceLimitsToAPI(s.ResourceLimits),
		WorkingDir:     s.WorkingDir,
		NetworkMode:    s.NetworkMode,
		Labels:         s.Labels,
		PullPolicy:     api.PullPolicy(s.PullPolicy),
		Privileged:     s.Privileged,
		TimeoutSeconds: s.TimeoutSeconds,
	}
}

func volumeMountsFromAPI(in []api.VolumeMount) []domain.VolumeMount {
	if in == nil {
		return nil
	}
	out := make([]domain.VolumeMount, len(in))
	for i, m := range in {
		out[i] = domain.VolumeMount{
			HostPath:      m.HostPath,
			ContainerPath: m.ContainerPath,
			ReadOnly:      m.ReadOnly,
		}
	}
	return out
}

func volumeMountsToAPI(in []domain.VolumeMount) []api.VolumeMount {
	if in == nil {
		return nil
	}
	out := make([]api.VolumeMount, len(in))
	for i, m := range in {
		out[i] = api.VolumeMount{
			HostPath:      m.HostPath,
			ContainerPath: m.ContainerPath,
			ReadOnly:      m.ReadOnly,
		}
	}
	return out
}

func resourceLimitsFromAPI(in *api.ResourceLimits) *domain.ResourceLimits {
	if in == nil {
		return nil
	}
	return &domain.ResourceLimits{
		MemoryBytes: in.MemoryBytes,
		CPUCores:    in.CPUCores,
	}
}

func resourceLimitsToAPI(in *domain.ResourceLimits) *api.ResourceLimits {
	if in == nil {
		return nil
	}
	return &api.ResourceLimits{
		MemoryBytes: in.MemoryBytes,
		CPUCores:    in.CPUCores,
	}
}
