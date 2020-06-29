package main

import (
	"fmt"
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
	//fmt.Println(c.RemoteAddr().String())
	ctx := http.AcquireContext(httpServer, conn)
	conn.SetContext(ctx)
	return nil
}

func (s *Server) OnMessage(conn *qio.Conn) error {
	ctx, ok := conn.GetContext().(*http.Context)
	if !ok {
		log.Fatal("something wrong")
	}
	if err := ctx.ServeHttp(); err != nil {
		conn.Close()
		return err
	}
	return nil
}

func (s *Server) OnClose(conn *qio.Conn) {
	if conn.GetContext() != nil {
		ctx := conn.GetContext().(*http.Context)
		http.ReleaseContext(ctx)
		fmt.Println(conn.RemoteAddr(), "close")
	}
}

func handler(ctx *http.Context) {
	//fmt.Println(*ctx.Request().Header())
	resp := ctx.Response()
	resp.SetBody([]byte("hello"))
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
	server.Serve("tcp", ":9999")
}
