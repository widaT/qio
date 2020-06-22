package qio

import (
	"net"

	"github.com/widaT/qio/buf"
	"golang.org/x/sys/unix"
)

type Conn struct {
	buf        *buf.LinkedBuffer
	fd         int
	localAddr  net.Addr // local addr
	remoteAddr net.Addr // remote addr
	context    interface{}
}

func NewConn(fd int, sa unix.Sockaddr) *Conn {
	c := new(Conn)
	c.buf = buf.New()
	c.fd = fd
	c.remoteAddr = Sockaddr2TCP(sa)
	return c
}

func (c *Conn) NexWritablePos() []byte {
	return c.buf.NexWritablePos()
}

func (c *Conn) SetContext(context interface{}) {
	c.context = context
}

func (c *Conn) GetContext() interface{} {
	return c.context
}

func (c *Conn) Bytes() []byte {
	return c.buf.Bytes()
}

func (c *Conn) Shift(n int) {
	c.buf.Shift(n)
}

func (c *Conn) MoveWritePiont(n int) {
	c.buf.MoveWritePiont(n)
}

func (c *Conn) BufferPoint() *buf.LinkedBuffer {
	return c.buf
}

func (c *Conn) Read(b []byte) (n int, e error) {
	n, e = c.buf.Read(b)
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	n, err = unix.Write(c.fd, b)
	if err != nil {
		if err == unix.EAGAIN {
			err = nil
			return
		}
	}
	return
}

func (c *Conn) Close() error {
	c.buf.Release()
	return unix.Close(c.fd)
}

func (c *Conn) LocalAddr() net.Addr {
	return c.localAddr

}
func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}
