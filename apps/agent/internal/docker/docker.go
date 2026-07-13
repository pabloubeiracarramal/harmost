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
