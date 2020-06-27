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

func (s *Server) OnMessage(conn *qio.Conn) error {
	var ctx *http.Context
	if conn.GetContext() == nil {
		ctx = http.AcquireContext(httpServer, conn)
		conn.SetContext(ctx)
	} else {
		ctx = conn.GetContext().(*http.Context)
	}
	if err := ctx.ServeHttp(); err != nil {
		fmt.Println(err)
		if err != http.StatusPartial {
			conn.Close()
			return err
		}
	}
	ctx.Reset(conn)
	return nil
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
	server.Serve("tcp", "127.0.0.1:9999")
}
