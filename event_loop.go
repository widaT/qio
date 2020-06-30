package qio

import (
	"log"
	"sync/atomic"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"

	"golang.org/x/sys/unix"
)

var evId uint32
var nextIdx uint32

type EventLoop struct {
	id          uint32
	server      *Server
	poller      *poller.Poller
	connections map[int]*Conn
	evServer    EventServer
}

func (e *EventLoop) close() {
	for _, conn := range e.connections {
		conn.Close()
	}
}

func (e *EventLoop) run() {
	defer e.close()
	log.Printf("ev %d exit err:%s", e.id, e.poller.Polling(e.handleEvent))
}

func (e *EventLoop) accept(fd int, sa unix.Sockaddr) error {
	var ev *EventLoop
	if e.server.portReuse {
		ev = e
	} else {
		if atomic.LoadUint32(&nextIdx) == uint32(len(e.server.subEventLoop)) {
			nextIdx = 0
		}
		ev = e.server.subEventLoop[nextIdx]
		atomic.AddUint32(&nextIdx, 1)
	}

	conn := NewConn(ev, fd, sa)
	ev.runTask(func() {
		ev.connections[fd] = conn
		err := ev.poller.Register(fd, ClientToken, interest.READABLE, pollopt.Level)
		if err != nil {
			log.Println(err)
			return
		}
		ev.evServer.OnConnect(conn)
	})

	return nil
}

func (e *EventLoop) runTask(fn func()) {
	e.poller.AddTask(fn)
	err := e.poller.Wake()
	if err != nil {
		log.Printf("%s", err)
	}
}

func (e *EventLoop) CloseConn(conn *Conn) {
	var err error
	delete(e.connections, conn.fd)
	err = e.poller.Deregister(conn.fd)
	if err != nil {
		log.Printf("%v", err)
	}
	e.evServer.OnClose(conn)
	log.Println(conn.remoteAddr.String(), "close")
	err = unix.Close(conn.fd)
	if err != nil {
		log.Printf("%v", err)
	}
	conn.buf.Release()
	conn.outbuf.Release()
}

func (e *EventLoop) handleEvent(ev *poller.Event) error {
	fd := int(ev.Fd)

	switch ev.Token() {
	case ServerToken:
		switch {
		case ev.IsReadable():
			cfd, sa, err := unix.Accept(fd)
			if err != nil {
				//WouldBlock
				if err == unix.EAGAIN {
					return nil
				}
				return err
			}
			if err := poller.Nonblock(cfd); err != nil {
				return err
			}

			//	SetKeepAlive(cfd, 3)
			return e.accept(cfd, sa)
		}
	case ClientToken:
		if conn, found := e.connections[fd]; found {
			var err error
			switch {
			case ev.IsReadClosed() || ev.IsError() || ev.IsWriteClosed():
				e.CloseConn(conn)
			case ev.IsReadable():
				connectionClosed := false
				for {
					b := conn.NexWritablePos()
					n, err := unix.Read(fd, b)
					if n == 0 {
						connectionClosed = true
						break
					}
					if err != nil {
						//WouldBlock
						if err == unix.EAGAIN {
							break
						}
						//Interrupted
						if err == unix.EINTR {
							continue
						}
						//防止 connection reset by peer 的情况下程序退出
						connectionClosed = true
						break
					}
					conn.MoveWritePiont(n)
				}
				if !connectionClosed {
					err = e.evServer.OnMessage(conn)
					if err != nil {
						connectionClosed = true
					}
				}
				if connectionClosed {
					e.CloseConn(conn)
				}
			case ev.IsWritable():
				if conn.outbuf.Buffered() == 0 {
					e.poller.Reregister(fd, ClientToken, interest.READABLE, pollopt.Level)
					return nil
				}
				b, n := conn.outbuf.Bytes()
				n, err := unix.Write(fd, b)
				if err != nil {
					if err == unix.EAGAIN {
						return nil
					}
					return err
				}
				conn.outbuf.Shift(n)
				if conn.outbuf.Buffered() == 0 {
					e.poller.Reregister(fd, ClientToken, interest.READABLE, pollopt.Level)
				}
			}
		}
	}
	return nil
}

func SetKeepAlive(fd, secs int) error {
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, secs); err != nil {
		return err
	}
	return unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, secs)
}
