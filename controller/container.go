package controller

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

func (c *Client) clearContainer(ctx context.Context) error {
	containers, err := c.dockerClient.ContainerList(
		ctx,
		types.ContainerListOptions{
			All: true,
		})
	if err != nil {
		return err
	}
	for _, ct := range containers {
		err = c.dockerClient.ContainerRemove(ctx, ct.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) createContainer(ctx context.Context, labels map[string]string, image string) (container.ContainerCreateCreatedBody, error) {
	body, err := c.dockerClient.ContainerCreate(ctx,
		&container.Config{
			Image:  image,
			Labels: labels,
		},
		nil, nil, "")
	return body, err
}