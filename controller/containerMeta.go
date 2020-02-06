package controller

import (
	"time"
)

type containerMeta struct {
	id string
	funcName string
	containerName string
	version int32
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
	for {
		select {
		case t := <- c.inTasks:
			c.waitedTasks[t.id] = t
			ib := &interactionPackage{
				interactionHeader{
					t.id,
					uint64(len(t.args)),
				},
				[]byte(t.args),
			}
			if err := c.conn.write(ib); err != nil {
				// TODO
				panic(err)
			}
		case r := <- c.outResponses:
			c.waitedTasks[r.id].res <- r.res
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
			c.outResponses <- &response{
				id: p.id,
				res: p.body,
			}
		}
	}
}