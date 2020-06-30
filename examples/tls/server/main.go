package main

import (
	"log"

	"github.com/widaT/qio"
	tls "github.com/widaT/tls13"
)

type Server struct {
	*qio.DefaultEvServer
}

var tlsConfig *tls.Config

func init() {
	c, err := tls.LoadX509KeyPair("examples/tls/pem/cert.pem", "examples/tls/pem/key.pem")
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig = &tls.Config{Certificates: []tls.Certificate{c}}
}

func (s *Server) OnConnect(conn *qio.Conn) error {
	tlsConn := tls.Server(conn, tlsConfig)
	conn.SetContext(tlsConn)
	return nil
}

func (s *Server) OnMessage(conn *qio.Conn) error {
	tlsConn := conn.GetContext().(*tls.Conn)
	b, n, err := tlsConn.ReadN(81)
	if tls.StatusPartial == err {
		return nil
	}
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("receive %s \n", b)
	tlsConn.Write(b)
	tlsConn.Shift(n)
	return nil
}

func main() {
	server, err := qio.NewServer(new(Server))
	if err != nil {
		log.Fatal(err)
	}
	server.Serve("tcp", ":9999")
}
