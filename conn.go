package qio

import (
	"net"
	"time"

	"golang.org/x/sys/unix"
)

//Conn shoud imp net.Conn
type Conn struct {
	buf        *LinkedBuffer
	fd         int
	localAddr  net.Addr // local addr
	remoteAddr net.Addr // remote addr
}

func NewConn(fd int, sa unix.Sockaddr) *Conn {
	conn := new(Conn)
	conn.buf = New()
	conn.fd = fd
	conn.remoteAddr = SockaddrToTCPOrUnixAddr(sa)
	return conn
}

/* func (conn *Conn) Buffered() []byte {
	length := conn.buf.Buffered()
	b, _ := conn.buf.ReadN(length)
	return b
} */

func (conn *Conn) Read(b []byte) (n int, e error) {
	n, e = conn.buf.Read(b)
	return
}

func (conn *Conn) Write(b []byte) (n int, err error) {
	//fmt.Printf("write %s\n", b)
	return unix.Write(conn.fd, b)
}

func (conn *Conn) Close() error {
	conn.buf.Release()
	return unix.Close(conn.fd)
}

func (conn *Conn) LocalAddr() net.Addr {
	return conn.localAddr

}

func (conn *Conn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}

func (conn *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (conn *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}
