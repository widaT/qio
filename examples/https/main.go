package main

import (
	"log"
	"syscall"

	http "github.com/widaT/http1"
	"github.com/widaT/qio"
	tls "github.com/widaT/tls13"
)

var httpServer *http.Server

var tlsConfig *tls.Config

func init() {
	c, err := tls.LoadX509KeyPair("examples/tls/pem/cert.pem", "examples/tls/pem/key.pem")
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig = &tls.Config{Certificates: []tls.Certificate{c}}
}

type Server struct {
	*qio.DefaultEvServer
}

func (s *Server) OnConnect(conn *qio.Conn) error {

	tlsConn := tls.Server(conn, tlsConfig)

	ctx := http.AcquireContext(httpServer, tlsConn)
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
		log.Println("----", err)
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
	server.Serve("tcp", ":9999")
}
