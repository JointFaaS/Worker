package controller

import (
	"context"
	"log"
)

type functionMeta struct {
	name string
	image string
}

type funcState int
const (
	running funcState = 0
	creating funcState = 1
	cold funcState = 2
)

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
				log.Print("Cold Function Request")
				c.funcStateMap[t.funcName] = cold
				c.containerMap[t.funcName] = make([]containerMeta, 0)
				c.subTasks[t.funcName] = make(chan *task, 100)
				fState = cold
			}

			if  fState == running {
				c.subTasks[t.funcName] <- t
			} else if fState == cold {
				log.Print("create container")
				c.funcStateMap[t.funcName] = running
				c.subTasks[t.funcName] <- t
				go func ()  {
					_, err := c.createContainer(context.TODO(), c.convertFuncNameToImageName(t.funcName))
					if err != nil {
						log.Print(err.Error())
					}
				}()
			}
		case ccr := <- c.createContainerResponse:
			ccr.inTasks = c.subTasks[ccr.funcName]
			c.containerMap[ccr.funcName] = append(c.containerMap[ccr.funcName], *ccr)
			go ccr.work()
		case <- ctx.Done():
			return
		}
	}
}