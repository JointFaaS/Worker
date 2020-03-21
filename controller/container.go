package controller

import (
	"context"
	"log"
	"time"
	"net"
	"bytes"
	"encoding/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

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

func (c *Client) createContainer(ctx context.Context, labels map[string]string, envs []string, image string, codeDir string) (container.ContainerCreateCreatedBody, error) {
	body, err := c.dockerClient.ContainerCreate(ctx, 
		&container.Config{
			Image: image,
			Env: envs,
			Labels: labels,
		},
		&container.HostConfig{
			Binds: []string{
				c.config.SocketPath + ":/var/run/worker.sock",
				codeDir + ":/tmp/code",
			},
		},
		nil, "")
	return body, err
}

func (c *Client) workForContainerRegistration() {
	for {
		unixConn, err := c.unixListener.AcceptUnix()
		if err != nil {
			continue
		}
		log.Printf("new connection")
		go func ()  {
			err := c.registerHelper(unixConn)
			if err != nil {
				log.Print(err.Error())
			}
		}() 
	}
}

type registerBody struct {
	FuncName string `json:"funcName"`
	EnvID string `json:"envID"`
}

func (c *Client) registerHelper(unixConn *net.UnixConn) error {
	cc := newContainerConn(unixConn)
	for {
		if err := cc.poll(time.Now().Add(time.Second)); err != nil {
			return err
		}
		ib, err := cc.read()
		if err != nil {
			return err
		}
		if ib != nil {
			var regBody registerBody
			err := json.NewDecoder(bytes.NewReader(ib.body)).Decode(&regBody)
			if err != nil {
				return err
			}
			c.containerRegistration <- newContainerMeta(regBody.EnvID, regBody.FuncName, cc)
			return nil
		}
	}
}