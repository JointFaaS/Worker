package controller

import (
	"container/list"
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

func (c *Client) addSpecifiedContainer(funcName string) error {
	c.resourceMu.RLock()
	resource, isPresent := c.funcResourceMap[funcName]
	c.resourceMu.RUnlock()
	if isPresent == false {
		return errors.New("Such Function has not been initialised")
	}

	if resource.Runtime == "custom" {
		c.addContainer(resource.Image, resource.MemorySize, resource.FuncName)
		return nil
	}

	c.idleContainerMu.Lock()
	idleContainers, isPresent := c.idleContainerMap[resource.MemorySize]
	if isPresent && idleContainers.Len() > 0 {
		c.containerMu.Lock()
		containers, cIsPresent := c.funcContainerMap[resource.FuncName]
		if cIsPresent == false {
			containers = list.New()
			c.funcContainerMap[resource.FuncName] = containers
		}
		e := idleContainers.Remove(idleContainers.Front())
		containers.PushBack(e)
		c.containerMu.Unlock()
		c.idleContainerMu.Unlock()
		go c.addGeneralContainer(resource.Image, resource.MemorySize)
		return nil
	}
	c.idleContainerMu.Unlock()
	c.addContainer(resource.Image, resource.MemorySize, resource.FuncName)
	return nil
}

func (c *Client) addContainer(image string, memorySize int64, funcName string) (string, error) {
	if funcName != "" {
		c.creatingContainerMu.Lock()
		num, isPresent := c.creatingContainerNumMap[funcName]
		if isPresent == false || num == 0 {
			c.creatingContainerNumMap[funcName] = 1
		} else {
			return "", nil
		}
		c.creatingContainerMu.Unlock()
	}

	container, err := c.dockerClient.ContainerCreate(context.TODO(),
		&dtc.Config{
			Image: image,
		},
		&dtc.HostConfig{},
		nil, "")
	if err != nil {
		return "", err
	}
	err = c.dockerClient.ContainerStart(context.TODO(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}
	return container.ID, nil
}

func (c *Client) addGeneralContainer(image string, memorySize int64) (string, error) {
	return c.addContainer(image, memorySize, "")
}
