package qio

import (
	"net"
	"time"

	"golang.org/x/sys/unix"
)

//Conn shoud imp net.Conn
type Conn struct {
	buf        *ConpositeBuf
	linkedBuf  *LinkedBuffer
	fd         int
	localAddr  net.Addr // local addr
	remoteAddr net.Addr // remote addr
}

func NewConn(fd int, sa unix.Sockaddr) *Conn {
	conn := new(Conn)
	conn.buf = &ConpositeBuf{}
	conn.linkedBuf = New()
	conn.fd = fd
	conn.remoteAddr = SockaddrToTCPOrUnixAddr(sa)
	return conn
}

func (conn *Conn) Read(b []byte) (n int, e error) {
	return conn.buf.Read(b)
}

func (conn *Conn) Write(b []byte) (n int, err error) {
	return unix.Write(conn.fd, b)
}

func (conn *Conn) Close() error {
	conn.buf.Drop()
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
