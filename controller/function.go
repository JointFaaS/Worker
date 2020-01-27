package controller

import (
	"context"
)

// FunctionMeta defines the prop of a function
type FunctionMeta struct {
	name string
	image string
}

// Invoke pass a function request to backend
func (c *Client) Invoke(ctx context.Context, name string, args string, res chan string)  {
	c.tasks <- &task{funcName: name, args: args, res: res, ctx: ctx}
}

func dispatch(container *ContainerMeta, t *task) {
	// TODO
}

func (c *Client) randomAvailableContainer(t *task) *ContainerMeta{
	// TODO
	return nil
}
func (c *Client) work() {
	for {
		select {
		case <- c.tasks:
			t := <- c.tasks
			fState, isPresent := c.funcStateMap[t.funcName]
			if isPresent == false {
				c.funcStateMap[t.funcName] = cold
				fState = cold
			}

			if  fState == running {
				availableContainer := c.randomAvailableContainer(t)
				go dispatch(availableContainer, t)
			} else if fState == creating {
				// TODO
			} else if fState == cold {
				c.funcStateMap[t.funcName] = creating
				// TODO
			}
		}
	}
}