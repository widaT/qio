package qio

import (
	"fmt"
	"log"

	"github.com/widaT/poller"
	"github.com/widaT/qio/conn"
	tls "github.com/widaT/qio/tls13"
	"golang.org/x/sys/unix"
)

var evId uint32

type EventLoop struct {
	id          uint32
	server      *Server
	poller      *poller.Selector
	connections map[int]conn.Conn
	handler     Handler
	tlsConfig   *tls.Config
}

func (e *EventLoop) close() {
	for _, conn := range e.connections {
		conn.Close()
	}
}

func (e *EventLoop) run() {
	defer e.close()
	log.Printf("ev %d exit err:%s", e.id, poller.Polling(e.poller, e.handleEvent))
}

func (e *EventLoop) handleEvent(ev *poller.Event) error {
	switch ev.Token() {
	case poller.Token(0):
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
			e.server.accept(cfd, sa)
		}
	case poller.Token(1):
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
				if ok {
					if !tConn.ConnectionState().HandshakeComplete {
						err := tConn.Handshake()
						if err != nil {
							connectionClosed = true
						}
					} else {
						err := e.handler(conn)
						if err != nil {
							delete(e.connections, int(ev.Fd))
							conn.Close()
						}
					}
				} else {
					err := e.handler(conn)
					if err != nil {
						delete(e.connections, int(ev.Fd))
						conn.Close()
					}
				}
				if connectionClosed {
					fmt.Println("connectionClosed")
					delete(e.connections, int(ev.Fd))
					conn.Close()
				}
			}
		}
	}
	return nil
}
