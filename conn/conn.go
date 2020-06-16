package conn

import (
	"net"
	"time"

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

//Conn shoud imp net.Conn
type conn struct {
	buf        *buf.LinkedBuffer
	fd         int
	localAddr  net.Addr // local addr
	remoteAddr net.Addr // remote addr
}

func NewConn(fd int, sa unix.Sockaddr) Conn {
	c := new(conn)
	c.buf = buf.New()
	c.fd = fd
	c.remoteAddr = SockaddrToTCPOrUnixAddr(sa)
	return c
}

func (c *conn) NexWriteBlock() []byte {
	return c.buf.NexWriteBlock()
}

func (c *conn) MoveWritePiont(n int) {
	c.buf.MoveWritePiont(n)
}
func (c *conn) BufferPoint() *buf.LinkedBuffer {
	return c.buf
}

func (c *conn) Read(b []byte) (n int, e error) {
	n, e = c.buf.Read(b)
	return
}

func (c *conn) Write(b []byte) (n int, err error) {
	return unix.Write(c.fd, b)
}

func (c *conn) Close() error {
	c.buf.Release()
	return unix.Close(c.fd)
}

func (c *conn) LocalAddr() net.Addr {
	return c.localAddr

}
func (c *conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return nil
}
