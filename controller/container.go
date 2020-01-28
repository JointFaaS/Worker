package controller

import (
	"context"
	"syscall"
	"os"
	"path"
	
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
)
// ContainerMeta defines the metadata of a container
type ContainerMeta struct {
	funcName string
	version int32
	sequenceID int32
}

func prepareNamedPipeForContainer(name string) {
	err := os.Mkdir(path.Join("/tmp", name), 0777)
	if err != nil {
		panic(err)
	}
	err = syscall.Mkfifo(path.Join("/tmp", name, "down") , 0666)
	if err != nil {
		panic(err)
	}
	err = syscall.Mkfifo(path.Join("/tmp", name, "up") , 0666)
	if err != nil {
		panic(err)
	}
}

func (c *Client) createContainer(ctx context.Context, containerName string, image string) (container.ContainerCreateCreatedBody, error) {
	prepareNamedPipeForContainer(containerName)

	body, err := c.dockerClient.ContainerCreate(ctx, 
		&container.Config{
			Image: image,
		},  
		&container.HostConfig{
			Mounts: []mount.Mount{
				mount.Mount{
					Type: mount.TypeNamedPipe,
					Source: path.Join("/tmp", containerName, "up"),
					Target: "/up",
				},
				mount.Mount{
					Type: mount.TypeNamedPipe,
					Source: path.Join("/tmp", containerName, "down"),
					Target: "/down",
				},
			},
			NetworkMode: "none",
		},
		nil, containerName)
	return body, err
}

func (c *Client) clearContainer(ctx context.Context) (error) {
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

type containerPipe struct {
	up *os.File
	down *os.File
}

// GetNamedPipeOfEnv returns the up and down pipes for a running container
func (c *Client) getNamedPipeOfContainer(containerName string) (*containerPipe, error){
	pipes, isPresent := c.containerPipeMap[containerName]
	if isPresent {
		return pipes, nil
	}
	up, err := os.OpenFile(path.Join("/tmp", containerName, "up"), os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	down, err := os.OpenFile(path.Join("/tmp", containerName, "down"), os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		up.Close()
		return nil, err
	}
	pipes = &containerPipe{up: up, down: down}
	c.containerPipeMap[containerName] = pipes
	return pipes, nil
}

