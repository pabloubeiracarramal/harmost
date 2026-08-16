package docker

// Main responsibility is wrapping the Docker SDK client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type Docker struct {
	cli *client.Client
}

func New() (*Docker, error) {

	cli, err := client.New(
		client.FromEnv,
	)

	if err != nil {
		return nil, err
	}

	return &Docker{cli: cli}, nil

}

func (d *Docker) Ping(ctx context.Context) error {
	_, err := d.cli.Ping(ctx, client.PingOptions{})
	return err
}

func (d *Docker) ListAllContainers(ctx context.Context) ([]container.Summary, error) {
	res, err := d.cli.ContainerList(ctx, client.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (d *Docker) StartContainer(ctx context.Context, id string) error {
	_, err := d.cli.ContainerStart(ctx, id, client.ContainerStartOptions{})
	return err
}

func (d *Docker) StopContainer(ctx context.Context, id string) error {
	_, err := d.cli.ContainerStop(ctx, id, client.ContainerStopOptions{})
	return err
}

func (d *Docker) RestartContainer(ctx context.Context, id string) error {
	_, err := d.cli.ContainerRestart(ctx, id, client.ContainerRestartOptions{})
	return err
}

// RemoveContainer does not force-remove: a running container is refused with
// Docker's own error rather than silently killed. The front only offers
// Remove once a container isn't running.
func (d *Docker) RemoveContainer(ctx context.Context, id string) error {
	_, err := d.cli.ContainerRemove(ctx, id, client.ContainerRemoveOptions{})
	return err
}

// ContainerStats returns a single live resource-usage sample for a running
// container. cpuPercent uses Docker's standard delta formula, which needs a
// previous sample — IncludePreviousSample makes the daemon collect two ~1s
// apart before responding, so this call takes slightly over a second.
func (d *Docker) ContainerStats(ctx context.Context, id string) (cpuPercent float64, memUsageBytes, memLimitBytes int64, err error) {
	res, err := d.cli.ContainerStats(ctx, id, client.ContainerStatsOptions{IncludePreviousSample: true})
	if err != nil {
		return 0, 0, 0, err
	}
	defer res.Body.Close()

	var stats container.StatsResponse
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return 0, 0, 0, fmt.Errorf("decode stats: %w", err)
	}

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	if systemDelta > 0 && cpuDelta > 0 {
		onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
		if onlineCPUs == 0 {
			onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}
		if onlineCPUs == 0 {
			onlineCPUs = 1
		}
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100
	}

	return cpuPercent, int64(stats.MemoryStats.Usage), int64(stats.MemoryStats.Limit), nil
}
