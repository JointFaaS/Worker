package controller

import (
	"context"
	"time"
)

type functionMeta struct {
	name string
	image string
}

// Invoke pass a function request to backend
func (c *Client) Invoke(ctx context.Context, name string, args string, res chan []byte)  {
	c.tasks <- &task{funcName: name, args: args, res: res, ctx: ctx}
}

func (c *Client) workForExternalRequest(ctx context.Context) {
	for {
		select {
		case t := <- c.tasks:
			fState, isPresent := c.funcStateMap[t.funcName]
			if isPresent == false {
				c.funcStateMap[t.funcName] = cold
				c.containerMap[t.funcName] = make([]containerMeta, 0)
				c.subTasks[t.funcName] = make(chan *task)
				fState = cold
			}

			if  fState == running {
				c.subTasks[t.funcName] <- t
			} else if fState == cold {
				c.funcStateMap[t.funcName] = running
				c.subTasks[t.funcName] <- t
				ctx, _ := context.WithTimeout(context.TODO(), time.Second * 10)
				go c.createContainer(ctx, c.convertFuncNameToImageName(t.funcName))
			}
		case ccr := <- c.createContainerResponse:
			c.containerMap[ccr.funcName] = append(c.containerMap[ccr.funcName], *ccr)
		case <- ctx.Done():
			return
		}
	}
}