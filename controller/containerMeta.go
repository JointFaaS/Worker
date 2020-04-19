package controller

import (
	"context"
	"log"
	"time"
)

type containerMeta struct {
	id string
	funcName string
	conn *containerConn
	waitedTasks map[uint64]*task
	inTasks chan *task
	outResponses chan *response
	concurrencyLimit int
	ctx context.Context
	cancel context.CancelFunc
}

type response struct {
	id uint64
	res []byte
}

func (c *containerMeta) workForInandOut() {
	var taskID uint64
	taskID = 0
	for {
		if len(c.waitedTasks) < c.concurrencyLimit {
			select {
			case t := <- c.inTasks:
				log.Printf("%s get inTask", c.id)
				c.waitedTasks[taskID] = t
				ib := &interactionPackage{
					interactionHeader{
						taskID,
						uint64(len(t.args)),
					},
					t.args,
				}
				taskID++
				if err := c.conn.write(ib); err != nil {
					// TODO
					c.inTasks <- t
					panic(err)
				}
			case r := <- c.outResponses:
				c.waitedTasks[r.id].res <- &Response{Err: nil, Body: &r.res}
				delete(c.waitedTasks, r.id)
			}
		}else {
			select {
			case r := <- c.outResponses:
				c.waitedTasks[r.id].res <- &Response{Err: nil, Body: &r.res}
				delete(c.waitedTasks, r.id)
			}
		}
	}
}

func (c *containerMeta) workForConnectionPoll() {
	for {
		p, err := c.conn.read()
		if err != nil {
			c.cancel()
			return
		}
		if p == nil {
			// poll
			err := c.conn.poll(time.Now().Add(time.Second))
			if err != nil {
				c.cancel()
				return
			}
		} else {
			log.Printf("%ul response", p.id)
			c.outResponses <- &response{
				id: p.id,
				res: p.body,
			}
		}
	}
}

func newContainerMeta(id string, funcName string, cc *containerConn) *containerMeta {
	return &containerMeta{
		id: id,
		funcName: funcName,
		conn: cc,
		waitedTasks: make(map[uint64]*task),
		outResponses: make(chan *response),
		// TODO
		concurrencyLimit: 1,
	}
}