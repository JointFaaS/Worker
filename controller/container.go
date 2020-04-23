package controller

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types"
	dtc "github.com/docker/docker/api/types/container"
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

func (c *Client) addSpeifiedFuncContainer(funcName string, targetNum int) error {
	c.resourceRWMu.RLock()
	resource, isPresent := c.funcResourceMap[funcName]
	c.resourceRWMu.RUnlock()
	if isPresent == false {
		return errors.New("Such Function has not been initialised")
	}
	go func() {
		c.containerMu.Lock()
		defer c.containerMu.Unlock()
		idleContainers, isPresent := c.idleContainerMap[resource.MemorySize]
		if isPresent == false {
			c.addIdleContainer(resource.Image, resource.MemorySize)
		}
		if resource.Runtime == "custom" {
			
		} else {
			for _, container := range idleContainers {
				if container.GetRuntime() == resource.Runtime {
					
				}
			}
		}
	}()
	return nil
}

func (c *Client) addIdleContainer(image string, memorySize int64) error {
	container, err := c.dockerClient.ContainerCreate(context.TODO(),
	&dtc.Config{
		Image:  image,
	},
	&dtc.HostConfig{

	},
	nil, "")
	if err != nil {
		return err
	}
	err = c.dockerClient.ContainerStart(context.TODO(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	return nil
}
