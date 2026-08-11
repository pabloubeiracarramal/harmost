package docker

// Main responsibility is wrapping the Docker SDK client

import (
	"context"

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

// ListRunningContainers is ListAllContainers filtered to State == "running",
// for the agent detail page's live containers view.
func (d *Docker) ListRunningContainers(ctx context.Context) ([]container.Summary, error) {
	all, err := d.ListAllContainers(ctx)
	if err != nil {
		return nil, err
	}
	running := make([]container.Summary, 0, len(all))
	for _, c := range all {
		if c.State == container.StateRunning {
			running = append(running, c)
		}
	}
	return running, nil
}
