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
func (c *Client) Invoke(ctx context.Context, name string, args string, res chan *Response)  {
	c.tasks <- &task{funcName: name, args: args, res: res, ctx: ctx}
}

// Init creates a congtainer env
func (c *Client) Init(ctx context.Context, name string, image string, codeURI string, res chan *Response) {
	c.initTasks <- &initTask{funcName: name, image: image, codeURI: codeURI, res: res, ctx: ctx}
}

func (c *Client) workForExternalRequest(ctx context.Context) {
	var idGenerator uint64
	idGenerator = 0
	for {
		select {
		case t := <- c.initTasks:
			_, isPresent := c.funcStateMap[t.funcName]
			if isPresent == true {
				t.res <- nil
				continue
			}
			log.Print("Init Function Request")
			fr, err := newFuncResource(t.funcName, t.image, t.codeURI)
			if err != nil {
				t.res <- &Response{Err: err, Body: nil}
				continue
			}
			c.funcResourceMap[t.funcName] = fr
			c.funcStateMap[t.funcName] = running
			c.containerMap[t.funcName] = make([]containerMeta, 0)
			c.subTasks[t.funcName] = make(chan *task, 100)
			log.Print("create container")
			go func ()  {
				body, err := c.createContainer(
					context.TODO(),
					map[string]string{"id": string(idGenerator)},
					[]string{"funcName="+t.funcName, "envID="+string(idGenerator)},
					fr.image,
					fr.sourceCodeDir)
				if err != nil {
					log.Print(err.Error())
					t.res <- nil
				} else {
					c.dockerClient.ContainerStart(context.TODO(), body.ID, types.ContainerStartOptions{})
					t.res <- &Response{Err: nil, Body: nil}
				}
			}()

		case t := <- c.tasks:
			// set unique id
			t.id = idGenerator
			idGenerator++
			fState, isPresent := c.funcStateMap[t.funcName]
			if isPresent == false {
				t.res <- nil
			}
			if fState == running {
				c.subTasks[t.funcName] <- t
			} else if fState == cold {
				// TODO
				t.res <- nil
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