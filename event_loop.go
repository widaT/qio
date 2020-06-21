package qio

import (
	"log"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"

	//	tls "github.com/widaT/qio/tls13"
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
	err := e.server.subEventLoop.poller.Register(fd, ClientToken, interest.READABLE, pollopt.Level)
	if err != nil {
		return err
	}
	conn := NewConn(fd, sa)
	e.server.subEventLoop.connections[fd] = conn
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
		for {
			cfd, sa, err := unix.Accept(int(ev.Fd))
			if err != nil {
				//WouldBlock
				if err == unix.EAGAIN {
					//	fmt.Println(err)
					break
				}
				return err
			}
			if err := poller.Nonblock(cfd); err != nil {
				return err
			}
			e.accept(cfd, sa)
		}
	case ClientToken:
		if conn, found := e.connections[int(ev.Fd)]; found {
			switch {
			case ev.IsReadable():
				connectionClosed := false
				for {
					b := conn.NexWritablePos()
					n, err := unix.Read(int(ev.Fd), b)
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
				err := e.evServer.OnMessage(conn)
				if err != nil {
					connectionClosed = true
				}
				if connectionClosed {
					log.Printf("conn %s connectionClosed err:%s", conn.RemoteAddr(), err)
					delete(e.connections, int(ev.Fd))
					e.evServer.OnClose(conn)
					conn.Close()
				}
			}
		}
	}
	return nil
}
