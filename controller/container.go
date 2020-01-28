package controller

import (
	"context"
	"net"
	"path"
	
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
)

type containerMeta struct {
	id string
	funcName string
	containerName string
	version int32
	conn net.Conn
	inCh chan []byte
}

func (c *containerMeta) work(){
	out := func ()  {
		b := *new([]byte)
		for {
			n, _ := c.conn.Read(b)
			if n != 0 {
				// TODO
			}
		}
	}
	in := func ()  {
		inMsg := <- c.inCh
		c.conn.Write(inMsg)
	}
	go out()
	go in()
}

func (c *Client) workForContainerInitialized(){
	for {
		_, err := c.unixListener.AcceptUnix()
		if err != nil {
			continue
		}
		// TODO
	}
}

func (c *Client) createContainer(ctx context.Context, containerName string, image string) (container.ContainerCreateCreatedBody, error) {
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

