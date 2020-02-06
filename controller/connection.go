package controller

import (
	"net"
	"time"
	"encoding/binary"
	"bytes"
	"log"
)

type interactionPackage struct {
	interactionHeader
	body []byte
}

type interactionHeader struct {
	id uint64
	length uint64
}

type connState int
const (
	waittingHeader connState = 0
	waittingBody connState = 1
	err 		connState = 2
)
type containerConn struct {
	conn net.Conn
	buf *bytes.Buffer
	state connState
	headerCache interactionHeader
	bufCache []byte
}
func (cc *containerConn) poll(t time.Time) (error) {
	if err := cc.conn.SetReadDeadline(t); err != nil {
		return nil
	}
	n, err := cc.conn.Read(cc.bufCache)
	if err != nil {
		return err
	}
	cc.buf.Write(cc.bufCache[:n])
	if err := cc.conn.SetReadDeadline(time.Time{}); err != nil {
		return nil
	}
	return nil
}

func (cc *containerConn) read() (*interactionPackage, error) {
	if cc.state == err {
		return nil, nil
	}
	if cc.state == waittingHeader {
		if cc.buf.Len() >= 16 {
			p := cc.buf.Next(16)
			cc.headerCache.id = binary.BigEndian.Uint64(p[:8])
			cc.headerCache.length = binary.BigEndian.Uint64(p[8:])
			cc.state = waittingBody
		}
	}
	if cc.state == waittingBody {
		if uint64(cc.buf.Len()) >= cc.headerCache.length {
			p := cc.buf.Next(int(cc.headerCache.length))
			log.Printf("body: %s", string(p))
			return &interactionPackage{
				cc.headerCache,
				p,
			}, nil
		}
	}

	return nil, nil
}

func (cc *containerConn) write(ib *interactionPackage) error {
	_, err := cc.conn.Write(ib.body)
	return err
}

func newContainerConn(nc net.Conn) *containerConn{
	cc := &containerConn{
		conn: nc,
		buf: bytes.NewBuffer(make([]byte, 0)),
		state: waittingHeader,
		bufCache: make([]byte, 4096),
	}
	return cc
}