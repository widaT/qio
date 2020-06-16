package conn

import (
	"net"

	"github.com/widaT/qio/buf"
	"golang.org/x/sys/unix"
)

type Conn interface {
	NexWriteBlock() []byte
	MoveWritePiont(n int)
	BufferPoint() *buf.LinkedBuffer
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

type QConn struct {
	buf        *buf.LinkedBuffer
	fd         int
	localAddr  net.Addr // local addr
	remoteAddr net.Addr // remote addr
}

func NewConn(fd int, sa unix.Sockaddr) Conn {
	c := new(QConn)
	c.buf = buf.New()
	c.fd = fd
	c.remoteAddr = Sockaddr2TCP(sa)
	return c
}

func (c *QConn) NexWriteBlock() []byte {
	return c.buf.NexWriteBlock()
}

func (c *QConn) MoveWritePiont(n int) {
	c.buf.MoveWritePiont(n)
}

func (c *QConn) BufferPoint() *buf.LinkedBuffer {
	return c.buf
}

func (c *QConn) Read(b []byte) (n int, e error) {
	n, e = c.buf.Read(b)
	return
}

func (c *QConn) Write(b []byte) (n int, err error) {
	return unix.Write(c.fd, b)
}

func (c *QConn) Close() error {
	c.buf.Release()
	return unix.Close(c.fd)
}

func (c *QConn) LocalAddr() net.Addr {
	return c.localAddr

}
func (c *QConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}
