package controller

import (
	"context"
	"syscall"
	"os"
	"path"
	
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)
// ContainerMeta defines the metadata of a container
type ContainerMeta struct {
	funcName string
	version int32
	sequenceID int32
}

func (c *Client) createContainer(ctx context.Context, name string, image string, sequence string) (container.ContainerCreateCreatedBody, error) {
	os.Mkdir(path.Join("/tmp", name), 0777)
	os.Mkdir(path.Join("/tmp", name, sequence), 0777)
	syscall.Mkfifo(path.Join("/tmp", name, sequence, "down") , 0666)
	syscall.Mkfifo(path.Join("/tmp", name, sequence, "up") , 0666)

	body, err := c.dockerClient.ContainerCreate(ctx, 
		&container.Config{
			Image: image,
		},  
		&container.HostConfig{
			Mounts: []mount.Mount{
				mount.Mount{
					Type: mount.TypeNamedPipe,
					Source: path.Join("/tmp", name, "up"),
					Target: "/up",
				},
				mount.Mount{
					Type: mount.TypeNamedPipe,
					Source: path.Join("/tmp", name, "down"),
					Target: "/down",
				},
			},
			NetworkMode: "none",
		},
		nil, name)
	return body, err
}