package qio

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/widaT/qio/conn"
	tls "github.com/widaT/qio/tls13"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"golang.org/x/sys/unix"
)

type Handler func(conn.Conn) error

type Server struct {
	poller *poller.Selector
	//ln   *listener
	handle Handler

	connections sync.Map
	tlsConfig   *tls.Config
}

func NewServer(hander Handler) (*Server, error) {
	server := new(Server)
	var err error
	server.poller, err = poller.New()
	if err != nil {
		return nil, err
	}
	server.handle = hander

	return server, nil
}

func (s *Server) ServeTLS(network, addr, cert, key string) {
	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Println(err)
		return
	}
	s.tlsConfig = &tls.Config{Certificates: []tls.Certificate{c}}
	s.Serve(network, addr)
}

func (s *Server) Serve(network string, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fd, err := poller.Listener2Fd(ln, true)
	if err != nil {
		return err
	}
	err = s.poller.Register(fd, poller.Token(0), interest.READABLE, pollopt.Edge)
	if err != nil {
		return err
	}
	fn := func(ev *poller.Event) error {

		/* 		defer func() {
			if err := recover(); err != nil {

				fmt.Printf("%s", err)
			}
		}() */

		switch ev.Token() {
		case poller.Token(0):
			for {
				cfd, sa, err := unix.Accept(fd)
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
				err = s.poller.Register(cfd, poller.Token(1), interest.READABLE.Add(interest.WRITABLE), pollopt.Level)
				if err != nil {
					return err
				}
				conn := conn.NewConn(cfd, sa)
				if s.tlsConfig != nil {
					s.connections.Store(cfd, tls.Server(conn, s.tlsConfig))
				} else {
					s.connections.Store(cfd, conn)
				}

			}
		case poller.Token(1):
			if connp, found := s.connections.Load(int(ev.Fd)); found {
				conn := connp.(conn.Conn)
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
							err = s.handle(conn)
							if err != nil {
								s.connections.Delete(fd)
								conn.Close()
							}
						}
					} else {
						err = s.handle(conn)
						if err != nil {
							s.connections.Delete(fd)
							conn.Close()
						}
					}
					if connectionClosed {
						fmt.Println("connectionClosed")
						s.connections.Delete(fd)
						conn.Close()
					}
				}
			}
		}
		return nil
	}
	poller.Polling(s.poller, fn)
	return nil
}
