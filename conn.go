package qio

import (
	"net"

	buf "github.com/widaT/linkedbuf"
	"golang.org/x/sys/unix"
)

type Conn struct {
	e          *EventLoop
	buf        *buf.LinkedBuffer
	outbuf     *buf.LinkedBuffer
	fd         int
	localAddr  net.Addr // local addr
	remoteAddr net.Addr // remote addr
	context    interface{}
}

func NewConn(ev *EventLoop, fd int, sa unix.Sockaddr) *Conn {
	c := new(Conn)
	c.e = ev
	c.buf = buf.New()

	//outbuf use lazy initialize
	//c.outbuf = buf.New()
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

func (c *Conn) Bytes() ([]byte, error) {
	b, _ := c.buf.Bytes()
	return b, nil
}

func (c *Conn) Buffered() int {
	return c.buf.Buffered()
}

func (c *Conn) ReadN(n int) ([]byte, int) {
	return c.buf.ReadN(n)
}

func (c *Conn) Shift(n int) {
	c.buf.Shift(n)
}

func (c *Conn) MoveWritePiont(n int) {
	c.buf.MoveWritePiont(n)
}

func (c *Conn) Read(b []byte) (n int, e error) {
	n, e = c.buf.Read(b)
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if c.outbuf == nil {
		c.outbuf = buf.New()
	}

	if c.outbuf.Buffered() != 0 {
		c.outbuf.Write(b)
		return len(b), nil
	}
	n, err = unix.Write(c.fd, b)
	if err != nil {
		if err == unix.EAGAIN {
			c.outbuf.Write(b)
			c.e.reRegisterReadWrite(c.fd, ClientToken)
			return
		}
		return
	}
	if n < len(b) {
		c.outbuf.Write(b[n:])
		c.e.reRegisterReadWrite(c.fd, ClientToken)
	}
	return len(b), nil
}

func (c *Conn) Close() error {
	c.e.runTask(func() {
		c.e.CloseConn(c)
	})
	return nil
}

func (c *Conn) LocalAddr() net.Addr {
	return c.localAddr

}
func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}
