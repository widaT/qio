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
)

var ServerToken = poller.NextToken()
var ClientToken = poller.NextToken()

type Server struct {
	poller *poller.Poller
	//ln   *listener
	evServer      EventServer
	mainEventLoop *EventLoop
	subEventLoop  *EventLoop
	tlsConfig     *tls.Config
}

func NewServer(evServer EventServer) (*Server, error) {
	server := new(Server)
	var err error
	server.poller, err = poller.NewPoller()
	if err != nil {
		return nil, err
	}
	server.evServer = evServer
	return server, nil
}

func (s *Server) newEventLoop() (e *EventLoop, err error) {
	e = new(EventLoop)
	e.poller, err = poller.NewPoller()
	if err != nil {
		return
	}
	e.id = atomic.AddUint32(&evId, 1) //so main eventloop id is 1,sub eventloop start 2
	e.connections = make(map[int]conn.Conn)
	e.server = s
	e.tlsConfig = s.tlsConfig
	e.evServer = s.evServer
	return
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
	s.mainEventLoop, err = s.newEventLoop()
	if err != nil {
		return err
	}

	err = s.mainEventLoop.poller.Register(fd, poller.Token(ServerToken), interest.READABLE, pollopt.Edge)
	if err != nil {
		return err
	}

	s.subEventLoop, err = s.newEventLoop()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		//runtime.LockOSThread()
		s.mainEventLoop.run()
		wg.Done()
	}()

	go func() {
		runtime.LockOSThread()
		s.subEventLoop.run()
		wg.Done()
	}()
	wg.Wait()
	return nil
}
