package controller

import (
	"context"
	"log"
	"time"
	"net"
	"encoding/binary"
	"io"
	"bytes"
	"sync"
	"encoding/json"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
)

type containerMeta struct {
	id string
	funcName string
	containerName string
	version int32
	conn net.Conn
	waitedTasks map[uint64]*task
	inTasks chan *task
	outResponses chan *response
}

type response struct {
	id uint64
	res []byte
}

func (c *containerMeta) work(){
	out := func ()  {
		b := make([]byte, 0)
		for {
			// ID (8bytes) res size (8bytes) res (var-len)
			n, _ := io.ReadAtLeast(c.conn, b, 16)
			id, _ := binary.Uvarint(b)
			resLen := binary.BigEndian.Uint64(b[8:])
			if (uint64)(n - 16) >= resLen {
				c.outResponses <- &response{
					id: id,
					res: b[16:],
				}
			}
		}
	}
	go out()
	// ID (8bytes) args size (8bytes) args (var-len)
	b := make([]byte, 16)
	for {
		select {
		case t := <- c.inTasks:
			size := uint64(len(t.args))
			binary.PutUvarint(b, t.id)
			binary.BigEndian.PutUint64(b[8:], size)
			c.conn.Write(b)
			c.conn.Write([]byte(t.args))
			c.waitedTasks[t.id] = t
		case r := <- c.outResponses:
			c.waitedTasks[r.id].res <- r.res
			delete(c.waitedTasks, r.id)
		}

	}
}

func (c *Client) createContainer(ctx context.Context, image string) (container.ContainerCreateCreatedBody, error) {
	body, err := c.dockerClient.ContainerCreate(ctx, 
		&container.Config{
			Image: image,
		}, 
		&container.HostConfig{
			Binds: []string{
				c.config.SocketPath + ":/var/run/worker.sock",
			},
			NetworkMode: "none",
		},
		nil, "")
	return body, err
}

func (c *Client) clearContainer(ctx context.Context) (error) {
	containers, err := c.dockerClient.ContainerList(
		ctx, 
		types.ContainerListOptions{
			All: true,
		})
	if err != nil {
		return err
	}
	for _, ct := range containers {
		err = c.dockerClient.ContainerRemove(ctx, ct.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) workForContainerRegistration() {
	for {
		unixConn, err := c.unixListener.AcceptUnix()
		if err != nil {
			continue
		}
		log.Printf("%s connected", unixConn.RemoteAddr().String())
		go func ()  {
			err := c.registerHelper(unixConn)
			if err != nil {
				log.Print(err.Error())
			}
		}() 
	}
}

type registerBody struct {
	funcName string
	envID string
}

func (c *Client) registerHelper(unixConn *net.UnixConn) error {
	b := make([]byte, 4096)
	buf := bytes.NewBuffer(make([]byte, 0))
	header := make([]byte, 16)
	var bodyLen uint64
	o := &sync.Once{}
	for {
		if err := unixConn.SetReadDeadline(time.Now().Add(time.Second*3)); err != nil {
			return err
		}
		n, err := unixConn.Read(b)
		if err != nil {
			break
		}
		buf.Write(b[:n])
		o.Do(func() {
			if buf.Len() >= 16 {
				buf.Read(header)
				bodyLen = binary.BigEndian.Uint64(header[8:16])
			}
		})

		if uint64(buf.Len()) >= bodyLen {
			break
		}
	}
	if err := unixConn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}
	var regBody registerBody
	err := json.NewDecoder(buf).Decode(&regBody)
	if err != nil {
		return err
	}
	log.Printf("%s register", regBody.funcName)
	c.containerRegistration <- &regBody
	return nil
}