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
	if e.server.settings.portReuse {
		ev = e
	} else {
		if atomic.LoadUint32(&nextIdx) == uint32(len(e.server.subEventLoop)) {
			nextIdx = 0
		}
		ev = e.server.subEventLoop[nextIdx]
		atomic.AddUint32(&nextIdx, 1)
	}

	conn := NewConn(ev, fd, sa)
	if e.server.settings.keepAlive {
		if err := setKeepAlive(fd); err != nil {
			return err
		}
		if err := setKeepAlivePeriod(fd, e.server.settings.keepAlivePeriod); err != nil {
			return err
		}
	}
	ev.runTask(func() {
		ev.connections[fd] = conn
		err := ev.registerRead(fd, ClientToken)
		if err != nil {
			log.Println(err)
			return
		}
		ev.evServer.OnConnect(conn)
	})

	return nil
}

func (e *EventLoop) registerRead(fd int, token poller.Token) error {
	return e.poller.Register(fd, token, interest.READABLE, pollopt.Level)
}

func (e *EventLoop) reRegisterRead(fd int, token poller.Token) error {
	return e.poller.Reregister(fd, token, interest.READABLE, pollopt.Level)
}

func (e *EventLoop) reRegisterReadWrite(fd int, token poller.Token) error {
	return e.poller.Reregister(fd, token, interest.READABLE.Add(interest.WRITABLE), pollopt.Level)
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
	//log.Println(conn.remoteAddr.String(), "close")
	err = unix.Close(conn.fd)
	if err != nil {
		log.Printf("%v", err)
	}
	conn.buf.Release()
	if conn.outbuf != nil {
		conn.outbuf.Release()
	}
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
				if conn.outbuf == nil || conn.outbuf.Buffered() == 0 {
					e.reRegisterRead(fd, ClientToken)
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
					e.reRegisterRead(fd, ClientToken)
				}
			}
		}
	}
	return nil
}

func setKeepAlive(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1)
}

func setKeepAlivePeriod(fd, secs int) (err error) {
	err = unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, secs)
	if err != nil {
		return err
	}
	err = unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, secs)
	return
}
