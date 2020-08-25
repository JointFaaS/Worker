package controller

import (
	"container/list"
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/docker/docker/api/types"
	dtc "github.com/docker/docker/api/types/container"
)

func (c *Client) ClearContainer(ctx context.Context) error {
	containers, err := c.dockerClient.ContainerList(
		ctx,
		types.ContainerListOptions{
			All: true,
		})
	if err != nil {
		return err
	}
	for _, ct := range containers {
		typ, isPresent := ct.Labels["type"]
		if isPresent && typ == "jointfaas" {
			err = c.dockerClient.ContainerRemove(ctx, ct.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				return err
			}
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
	_, err := c.addContainer(resource.Image, resource.MemorySize, resource.FuncName)
	return err
}

func (c *Client) addContainer(image string, memorySize int64, funcName string) (string, error) {
	if funcName != "" {
		c.creatingContainerMu.Lock()
		num, isPresent := c.creatingContainerNumMap[funcName]
		if isPresent == false || num == 0 {
			c.creatingContainerNumMap[funcName] = 1
			c.creatingContainerMu.Unlock()
		} else {
			c.creatingContainerMu.Unlock()
			return "", nil
		}
	}

	container, err := c.dockerClient.ContainerCreate(context.TODO(),
		&dtc.Config{
			Image: image,
			Env: []string{"WORK_HOST=" + c.localhost, "MEMORY=" + strconv.FormatInt(memorySize, 10), "FUNC_NAME=" + funcName},
			Labels: map[string]string{"type": "jointfaas"},
		},
		&dtc.HostConfig{
			Resources: dtc.Resources{
				Memory: memorySize * 1024 * 1024,
			},
		}, nil, "")

	log.Printf("[liu] container %v create with no err\n", container.ID)

	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	err = c.dockerClient.ContainerStart(context.TODO(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	log.Printf("[liu] container start with no err\n")
	return container.ID, nil
}

func (c *Client) addGeneralContainer(image string, memorySize int64) (string, error) {
	return c.addContainer(image, memorySize, "")
}
