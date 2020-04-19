package controller

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/google/uuid"
)

type functionMeta struct {
	name  string
	image string
}

type funcState int

const (
	running  funcState = 0
	creating funcState = 1
	cold     funcState = 2
)

// Invoke pass a function request to backend
func (c *Client) Invoke(ctx context.Context, name string, args []byte, res chan *Response) {
	c.tasks <- &task{funcName: name, args: args, res: res, ctx: ctx}
}

// Init creates a congtainer env
func (c *Client) Init(ctx context.Context, name string, image string, codeURI string, res chan *Response) {
	c.initTasks <- &initTask{funcName: name, image: image, codeURI: codeURI, res: res, ctx: ctx}
}

func (c *Client) workForExternalRequest() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		select {
		case t := <-c.initTasks:
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
			c.funcContainerMap[t.funcName] = make([]*containerMeta, 0)
			c.subTasks[t.funcName] = make(chan *task, 100)

			// ensure there are two positions for funcName and envID
			if cap(c.config.ContainerEnvVariables) - len(c.config.ContainerEnvVariables) < 2 {
				extend := make([]string, len(c.config.ContainerEnvVariables) + 2)
				copy(extend, c.config.ContainerEnvVariables)
				c.config.ContainerEnvVariables = extend
			}
			go func(envID string) {
				ctx, cancel := context.WithTimeout(c.ctx, time.Second * 3)
				defer cancel()

				body, err := c.createContainer(
					ctx,
					map[string]string{"funcName": t.funcName, "envID": envID},
					append(c.config.ContainerEnvVariables, "funcName="+t.funcName, "envID="+envID),
					fr.image,
					fr.sourceCodeDir)
				if err != nil {
					log.Print(err.Error())
					t.res <- &Response{Err: err}
				} else {
					c.containerIDMap.Store(envID, body.ID)
					c.dockerClient.ContainerStart(ctx, body.ID, types.ContainerStartOptions{})
					t.res <- &Response{Err: nil}
				}
			}(uuid.New().String())

		case t := <-c.tasks:
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
		case ccr := <-c.containerRegistration:
			log.Printf("%s start working for %s", ccr.id, ccr.funcName)
			ccr.inTasks = c.subTasks[ccr.funcName]
			c.funcContainerMap[ccr.funcName] = append(c.funcContainerMap[ccr.funcName], ccr)
			go ccr.workForInandOut()
			go ccr.workForConnectionPoll()
		case <-c.ctx.Done():
			log.Print("Controller Exits")
			return
		}
	}
}
