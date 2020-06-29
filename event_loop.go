package qio

import (
	"log"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"

	"golang.org/x/sys/unix"
)

var evId uint32

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
	ev := e.server.subEventLoop
	if e.server.portReuse {
		ev = e
	}
	err := ev.poller.Register(fd, ClientToken, interest.READABLE, pollopt.Level)
	if err != nil {
		return err
	}
	conn := NewConn(ev, fd, sa)
	ev.runTask(func() {
		ev.connections[fd] = conn
	})
	e.evServer.OnConnect(conn)
	return nil
}

func (e *EventLoop) runTask(fn func()) {
	e.poller.AddTask(fn)
	err := e.poller.Wake()
	if err != nil {
		log.Printf("%s", err)
	}
}

func (e *EventLoop) handleEvent(ev *poller.Event) error {
	switch ev.Token() {
	case ServerToken:
		cfd, sa, err := unix.Accept(int(ev.Fd))
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
		return e.accept(cfd, sa)
	case ClientToken:
		if conn, found := e.connections[int(ev.Fd)]; found {
			var err error
			switch {
			case ev.IsReadable():
				connectionClosed := false
				for {
					b := conn.NexWritablePos()
					n, err := unix.Read(int(ev.Fd), b)
					//fmt.Println("read ", n)
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
						log.Println(err)
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
					log.Printf("conn %s connectionClosed err:%v", conn.RemoteAddr(), err)
					delete(e.connections, int(ev.Fd))
					e.poller.Deregister(int(ev.Fd))
					e.evServer.OnClose(conn)
					conn.Close()
				}
			case ev.IsWritable():
				if conn.outbuf.Buffered() == 0 {
					e.poller.Reregister(int(ev.Fd), ClientToken, interest.READABLE, pollopt.Level)
					return nil
				}
				b, n := conn.outbuf.Bytes()
				n, err := unix.Write(int(ev.Fd), b)
				if err != nil {
					if err == unix.EAGAIN {
						return nil
					}
					return err
				}
				conn.outbuf.Shift(n)
				if conn.outbuf.Buffered() == 0 {
					e.poller.Reregister(int(ev.Fd), ClientToken, interest.READABLE, pollopt.Level)
				}
			}
		}
	}
	return nil
}
