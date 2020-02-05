package controller

import (
	"net"
	"bytes"
)

type interactionPackage struct {
	id uint64
	length uint64
	body []byte
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
}

func (cc *containerConn) read() (*interactionPackage, error) {
	if cc.state == waittingHeader {

	} else if cc.state == waittingBody {

	} else if cc.state == err {
		return nil, nil
	}
	ib := &interactionPackage{}
	return ib, nil
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
	}
	return cc
}