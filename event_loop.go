package qio

import (
	"log"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"github.com/widaT/qio/conn"
	tls "github.com/widaT/qio/tls13"
	"golang.org/x/sys/unix"
)

var evId uint32

type EventLoop struct {
	id          uint32
	server      *Server
	poller      *poller.Poller
	connections map[int]conn.Conn
	evServer    EventServer
	tlsConfig   *tls.Config
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
	err := e.server.subEventLoop.poller.Register(fd, ClientToken, interest.READABLE.Add(interest.WRITABLE), pollopt.Level)
	if err != nil {
		return err
	}
	conn := conn.NewConn(fd, sa)
	if e.tlsConfig != nil {
		conn = tls.Server(conn, e.tlsConfig)
	}
	e.server.subEventLoop.connections[fd] = conn
	e.evServer.OnContect(conn)
	return nil
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
					b := conn.NexWriteBlock()
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
					//info.conn.buf.Wrap(seg)
				}
				tConn, ok := conn.(*tls.Conn)
				var err error
				if ok {
					if !tConn.ConnectionState().HandshakeComplete {
						err = tConn.Handshake()
						if err != nil {
							connectionClosed = true
						}
					} else {
						err = e.evServer.OnMessage(conn)
						if err != nil {
							connectionClosed = true
						}
					}
				} else {
					err = e.evServer.OnMessage(conn)
					if err != nil {
						connectionClosed = true
					}
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
