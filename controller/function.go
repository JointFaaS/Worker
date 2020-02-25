package controller

import (
	"context"
	"log"
	"github.com/docker/docker/api/types"
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
	var idGenerator uint64
	idGenerator = 0
	for {
		select {
		case t := <- c.tasks:
			// set unique id
			t.id = idGenerator
			idGenerator++
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
					body, err := c.createContainer(
						context.TODO(),
						map[string]string{"id": string(idGenerator)},
						[]string{"funcName="+t.funcName, "envID="+string(idGenerator)},
						c.convertFuncNameToImageName(t.funcName))
					if err != nil {
						log.Print(err.Error())
					} else {
						c.dockerClient.ContainerStart(context.TODO(), body.ID, types.ContainerStartOptions{})
					}

				}()
			}
		case ccr := <- c.containerRegistration:
			log.Printf("%s start working", ccr.id)
			ccr.inTasks = c.subTasks[ccr.funcName]
			c.containerMap[ccr.funcName] = append(c.containerMap[ccr.funcName], *ccr)
			go ccr.workForIn()
			go ccr.workForOut()
		case <- ctx.Done():
			return
		}
	}
}