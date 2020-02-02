package controller

import (
	"context"
	"math/rand"
)

type functionMeta struct {
	name string
	image string
}

// Invoke pass a function request to backend
func (c *Client) Invoke(ctx context.Context, name string, args string, res chan string)  {
	c.tasks <- &task{funcName: name, args: args, res: res, ctx: ctx}
}

func (c *Client) dispatch(container *containerMeta, t *task) {
	
}

func (c *Client) create(t *task) {
	
}

func (c *Client) randomAvailableContainer(t *task) *containerMeta{
	containers, isPresent := c.containerMap[t.funcName]
	if isPresent == false {
		return nil
	}
	size := len(containers)
	if size == 0 {
		return nil
	}
	return &containers[rand.Intn(size)]
}

func (c *Client) work(ctx context.Context) {
	for {
		select {
		case t := <- c.tasks:
			fState, isPresent := c.funcStateMap[t.funcName]
			if isPresent == false {
				c.funcStateMap[t.funcName] = cold
				c.containerMap[t.funcName] = make([]containerMeta, 0)
				fState = cold
			}

			if  fState == running {
				availableContainer := c.randomAvailableContainer(t)
				go c.dispatch(availableContainer, t)
			} else if fState == creating {
				go c.create(t)
			} else if fState == cold {
				c.funcStateMap[t.funcName] = creating
				// TODO
			}
		case ccr := <- c.createContainerResponse:
			c.funcStateMap[ccr.funcName] = running
			c.containerMap[ccr.funcName] = append(c.containerMap[ccr.funcName], *ccr)
		case <- ctx.Done():
			return
		}
	}
}