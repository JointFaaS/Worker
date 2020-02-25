package controller

import (
	"time"
	"log"
)

type containerMeta struct {
	id string
	funcName string
	conn *containerConn
	waitedTasks map[uint64]*task
	inTasks chan *task
	outResponses chan *response
}

type response struct {
	id uint64
	res []byte
}

func (c *containerMeta) workForIn() {
	var taskID uint64
	taskID = 0
	for {
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
				panic(err)
			}
		case r := <- c.outResponses:
			c.waitedTasks[r.id].res <- &Response{Err: nil, Body: &r.res}
			delete(c.waitedTasks, r.id)
		}
	}
}

func (c *containerMeta) workForOut() {
	for {
		p, err := c.conn.read()
		if err != nil {
			// TODO
			panic(err)
		}
		if p == nil {
			// poll
			c.conn.poll(time.Now().Add(time.Second))
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
	}
}