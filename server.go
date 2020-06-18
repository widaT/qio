package qio

import (
	"log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"github.com/widaT/qio/conn"
	tls "github.com/widaT/qio/tls13"
	"golang.org/x/sys/unix"
)

type Handler func(conn.Conn) error

type Server struct {
	poller *poller.Selector
	//ln   *listener
	handle    Handler
	mainEl    *EventLoop
	subEl     *EventLoop
	tlsConfig *tls.Config
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

func (s *Server) newEventLoop() (e *EventLoop, err error) {
	e = new(EventLoop)
	e.poller, err = poller.New()
	if err != nil {
		return
	}
	e.id = atomic.AddUint32(&evId, 1) //so main eventloop id is 1,sub eventloop start 2
	e.connections = make(map[int]conn.Conn)
	e.server = s
	e.tlsConfig = s.tlsConfig
	e.handler = s.handle
	return
}

func (s *Server) accept(fd int, sa unix.Sockaddr) error {
	err := s.subEl.poller.Register(fd, poller.Token(1), interest.READABLE.Add(interest.WRITABLE), pollopt.Level)
	if err != nil {
		return err
	}
	conn := conn.NewConn(fd, sa)
	if s.tlsConfig != nil {
		conn = tls.Server(conn, s.tlsConfig)
	}
	s.subEl.connections[fd] = conn
	return nil
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
	s.mainEl, err = s.newEventLoop()
	if err != nil {
		return err
	}

	err = s.mainEl.poller.Register(fd, poller.Token(0), interest.READABLE, pollopt.Edge)
	if err != nil {
		return err
	}

	s.subEl, err = s.newEventLoop()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		runtime.LockOSThread()
		s.mainEl.run()
		wg.Done()
	}()

	go func() {
		runtime.LockOSThread()
		s.subEl.run()
		wg.Done()
	}()
	wg.Wait()

	return nil
}
