package controller

import (
	"context"
	"errors"
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
func (c *Client) Invoke(ctx context.Context, name string, args []byte, res chan *Response)  {
	c.tasks <- &task{funcName: name, args: args, res: res, ctx: ctx}
}

// Init creates a congtainer env
func (c *Client) Init(ctx context.Context, name string, image string, codeURI string, res chan *Response) {
	c.initTasks <- &initTask{funcName: name, image: image, codeURI: codeURI, res: res, ctx: ctx}
}

func (c *Client) workForExternalRequest(ctx context.Context) {
	for {
		select {
		case t := <- c.initTasks:
			_, isPresent := c.funcStateMap[t.funcName]
			if isPresent == true {
				t.res <- &Response{Err: errors.New("Init Repeatly")}
				continue
			}
			log.Print("Init Function Request")
			fr, err := newFuncResource(t.funcName, t.image, t.codeURI)
			if err != nil {
				t.res <- &Response{Err: err}
				continue
			}
			c.funcResourceMap[t.funcName] = fr
			c.funcStateMap[t.funcName] = running
			c.containerMap[t.funcName] = make([]*containerMeta, 0)
			c.subTasks[t.funcName] = make(chan *task, 100)
			log.Print("create container")
			go func ()  {
				body, err := c.createContainer(
					context.TODO(),
					map[string]string{"funcName": t.funcName},
					c.config.ContainerEnvVariables,
					fr.image,
					fr.sourceCodeDir)
				if err != nil {
					log.Print(err.Error())
					t.res <- &Response{Err: err}
				} else {
					c.dockerClient.ContainerStart(context.TODO(), body.ID, types.ContainerStartOptions{})
					t.res <- &Response{Err: nil, Body: nil}
				}
			}()

		case t := <- c.tasks:
			log.Printf("%s invoke", t.funcName)
			fState, isPresent := c.funcStateMap[t.funcName]
			if isPresent == false {
				t.res <- &Response{Err: errors.New("Not init function")}
				continue
			}
			if fState == running {
				c.subTasks[t.funcName] <- t
			} else if fState == cold {
				// TODO
				t.res <- &Response{Err: errors.New("Todo Cold")}
			}
		case ccr := <- c.containerRegistration:
			log.Printf("%s start working for %s", ccr.id, ccr.funcName)
			ccr.inTasks = c.subTasks[ccr.funcName]
			c.containerMap[ccr.funcName] = append(c.containerMap[ccr.funcName], ccr)
			go ccr.workForInandOut()
			go ccr.workForConnectionPoll()
		case <- ctx.Done():
			log.Print("Controller Exits")
			return
		}
	}
}