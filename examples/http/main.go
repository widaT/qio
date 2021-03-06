package main

import (
	"log"
	"syscall"

	http "github.com/widaT/http1"
	"github.com/widaT/qio"
)

var httpServer *http.Server

type Server struct {
	*qio.DefaultEvServer
}

func (s *Server) OnConnect(conn *qio.Conn) error {
	ctx := http.AcquireContext(httpServer, conn)
	conn.SetContext(ctx)
	//fmt.Println(conn.RemoteAddr(), "connect")
	return nil
}

func (s *Server) OnMessage(conn *qio.Conn) error {
	ctx, ok := conn.GetContext().(*http.Context)
	if !ok {
		log.Fatal("something wrong")
	}
	if err := ctx.ServeHttp(); err != nil {
		//here don't call conn.Close return err close conn int eventloop
		log.Println(err)
		return err
	}
	return nil
}

func (s *Server) OnClose(conn *qio.Conn) {
	if conn.GetContext() != nil {
		ctx := conn.GetContext().(*http.Context)
		http.ReleaseContext(ctx)
		//	fmt.Println(conn.RemoteAddr(), "close")
	}
}

func handler(ctx *http.Context) {
	//fmt.Println(*ctx.Request().Header())
	resp := ctx.Response()
	resp.SetBody([]byte("hello" + ctx.RemoteAddr().String()))
}

func main() {

	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	httpServer = http.NewServer(handler, 0)
	server, err := qio.NewServer(new(Server))
	if err != nil {
		log.Fatal(err)
	}
	server.Serve("tcp", ":9999", qio.SetKeepAlive(3)) //, qio.SetEventLoopNum(1))
}
