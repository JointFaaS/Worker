package controller

import (
	"context"
	"net"
	"encoding/binary"
	"io"

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
	waitedTasks map[uint64]*task
	inTasks chan *task
	outResponses chan *response
}

type response struct {
	id uint64
	res []byte
}

func (c *containerMeta) work(){
	out := func ()  {
		b := make([]byte, 0)
		for {
			// ID (8bytes) res size (8bytes) res (var-len)
			n, _ := io.ReadAtLeast(c.conn, b, 16)
			id, _ := binary.Uvarint(b)
			resLen, _ := binary.Uvarint(b[8:])
			if (uint64)(n - 16) >= resLen {
				c.outResponses <- &response{
					id: id,
					res: b[16:],
				}
			}
		}
	}
	go out()
	// ID (8bytes) args size (8bytes) args (var-len)
	b := make([]byte, 8)
	for {
		select {
		case t := <- c.inTasks:
			size := uint64(len(t.args))
			binary.PutUvarint(b, t.id)
			binary.PutUvarint(b[4:], size)
			c.conn.Write(b)
			c.conn.Write([]byte(t.args))
			c.waitedTasks[t.id] = t
		case r := <- c.outResponses:
			c.waitedTasks[r.id].res <- r.res
			delete(c.waitedTasks, r.id)
		}

	}
}

func (c *Client) createContainer(ctx context.Context, image string) (container.ContainerCreateCreatedBody, error) {
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
		nil, "")
	
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

