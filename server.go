package qio

import (
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/widaT/poller"
	"golang.org/x/sys/unix"
)

var ServerToken = poller.NextToken()
var ClientToken = poller.NextToken()

type Server struct {
	settings      *Settings
	poller        *poller.Poller
	evServer      EventServer
	mainEventLoop *EventLoop
	subEventLoop  []*EventLoop
	ln            net.Listener
	fd            int
	file          *os.File
}

func NewServer(evServer EventServer) (*Server, error) {
	server := new(Server)
	var err error
	server.poller, err = poller.NewPoller()
	if err != nil {
		return nil, err
	}
	//	server.portReuse = reuseSuported()
	server.evServer = evServer
	server.settings = defaultSetting()
	return server, nil
}

func (s *Server) newEventLoop() (e *EventLoop, err error) {
	e = new(EventLoop)
	e.poller, err = poller.NewPoller()
	if err != nil {
		return
	}
	e.id = atomic.AddUint32(&evId, 1) //so main eventloop id is 1,sub eventloop start 2
	e.connections = make(map[int]*Conn)
	e.server = s
	e.evServer = s.evServer
	return
}

func (s *Server) Listener2Fd(ln net.Listener, nonblock bool) (err error) {
	switch l := ln.(type) {
	case *net.TCPListener:
		s.file, err = l.File()
	case *net.UnixListener:
		s.file, err = l.File()
	}
	if err != nil {
		return
	}
	s.fd = int(s.file.Fd())
	if nonblock {
		err = unix.SetNonblock(s.fd, true)
	}
	return
}

func (s *Server) Serve(network string, addr string, opts ...Option) error {
	var err error
	s.ln, err = net.Listen(network, addr)
	if err != nil {
		return err
	}
	err = s.Listener2Fd(s.ln, true)
	if err != nil {
		return err
	}
	for _, o := range opts {
		o(s.settings)
	}
	if s.settings.portReuse {
		return s.runLoopsMode()
	}
	return s.runMainSubMode()
}

func (s *Server) runLoopsMode() (err error) {
	n := s.settings.eventLoopNum
	eventLoops := make([]*EventLoop, n)
	for i := 0; i < n; i++ {
		eventLoops[i], err = s.newEventLoop()
		if err != nil {
			return err
		}
		eventLoops[i].registerRead(s.fd, poller.Token(ServerToken))
	}
	wg := sync.WaitGroup{}
	for _, e := range eventLoops {
		wg.Add(1)
		go func(e *EventLoop) {
			runtime.LockOSThread()
			e.run()
			wg.Done()
		}(e)
	}
	wg.Wait()
	return nil
}

func (s *Server) runMainSubMode() (err error) {
	s.mainEventLoop, err = s.newEventLoop()
	if err != nil {
		return err
	}
	err = s.mainEventLoop.registerRead(s.fd, poller.Token(ServerToken))
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	n := s.settings.eventLoopNum
	s.subEventLoop = make([]*EventLoop, n)
	for i := 0; i < n; i++ {
		s.subEventLoop[i], err = s.newEventLoop()
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(i int) {
			runtime.LockOSThread()
			s.subEventLoop[i].run()
			wg.Done()
		}(i)
	}
	go func() {
		s.mainEventLoop.run()
		wg.Done()
	}()
	wg.Wait()
	return
}
