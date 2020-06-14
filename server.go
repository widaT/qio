package qio

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/widaT/qio/tls"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"golang.org/x/sys/unix"
)

type Handler func(net.Conn) error
type Info struct {
	nConn     net.Conn
	handshake bool
	conn      *Conn
	opened    bool
	skip      int
}

var BigBuf = make([]byte, 0x10000)

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
				err = s.poller.Register(cfd, poller.Token(1), interest.READABLE.Add(interest.WRITABLE), pollopt.Edge)
				if err != nil {
					return err
				}
				conn := NewConn(cfd, sa)
				if s.tlsConfig != nil {
					s.connections.Store(cfd, &Info{nConn: tls.Server(conn, s.tlsConfig), conn: conn})

				} else {
					s.connections.Store(cfd, &Info{nConn: conn, conn: conn})
				}

			}
		case poller.Token(1):
			if infoP, found := s.connections.Load(int(ev.Fd)); found {
				switch {
				case ev.IsReadable():
					connectionClosed := false
					for {
						//b := info.conn.linkedBuf.NexWriteBlock()
						n, err := unix.Read(int(ev.Fd), BigBuf)

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
							return err
						}
						info := infoP.(*Info)
						n, err = info.conn.buf.Write(BigBuf[:n])
						fmt.Println("wrote ", n)
						//seg := info.conn.linkedBuf.MoveWritePiont(n)
						//fmt.Printf("%s", seg.Byte())
						//info.conn.buf.Wrap(seg)
						/* 	if !info.opened {
						go func() { */
						/* 	err = s.handle(info.nConn)
						if err != nil {
							if err != tls.ErrPending {
								s.connections.Delete(fd)
								info.nConn.Close()
							}

						} */

						if !info.opened {
							//	info.opened = true

							tConn, ok := info.nConn.(*tls.Conn)
							if ok && !tConn.ConnectionState().HandshakeComplete {
								err := tConn.Handshake()
								if err != nil {
									//s.connections.Delete(fd)
									//info.nConn.Close()
									fmt.Println(err)
									continue
								}
								//info.handshake = true
							} else {
								info.handshake = true
							}
						}
						if info.handshake {
							err = s.handle(info.nConn)
							if err != nil {
								s.connections.Delete(fd)
								info.nConn.Close()
							}

						}
					}
					if connectionClosed {
						s.connections.Delete(fd)
						infoP.(*Info).nConn.Close()
					}
				}
			}

		}
		return nil
	}
	poller.Polling(s.poller, fn)
	return nil
}
