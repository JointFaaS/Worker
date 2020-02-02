package controller

import (
	"context"
	"net"
	"encoding/binary"

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
	ctx context.Context
	cancel func()
}

func (c *containerMeta) work(funcName string, inTasks chan *task){
	c.funcName = funcName
	if c.ctx != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(context.TODO())
	out := func ()  {
		b := make([]byte, 0)
		for {
			n, _ := c.conn.Read(b)
			if n != 0 {
				// TODO
			}
		}
	}
	in := func ()  {
		// callID (8bytes) args size (8bytes) args (var-len)
		b := make([]byte, 8)
		for {
			t := <- inTasks
			size := uint64(len(t.args))
			binary.PutUvarint(b, t.id)
			binary.PutUvarint(b[4:], size)
			c.conn.Write(b)
			c.conn.Write([]byte(t.args))
		}
	}

	go out()
	go in()
}

func (c *Client) createContainer(ctx context.Context, containerName string, image string) (container.ContainerCreateCreatedBody, error) {
	body, err := c.dockerClient.ContainerCreate(ctx, 
		&container.Config{
			Image: image,
		},  
		&container.HostConfig{
			Mounts: []mount.Mount{
				mount.Mount{
					Type: mount.TypeBind,
					Source: "/var/run/worker.sock",
					Target: "/var/run/worker.sock",
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

