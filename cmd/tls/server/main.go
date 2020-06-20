package main

import (
	"fmt"
	"log"

	"github.com/widaT/qio"
	"github.com/widaT/qio/conn"
	tls "github.com/widaT/tls13"
)

type Server struct {
	*qio.DefaultEvServer
}

var tlsConfig *tls.Config

func init() {
	c, err := tls.LoadX509KeyPair("tlsf/cert.pem", "tlsf/key.pem")
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig = &tls.Config{Certificates: []tls.Certificate{c}}
}

func (s *Server) OnContect(conn *conn.Conn) error {
	tlsConn := tls.Server(conn, tlsConfig)
	conn.SetContext(tlsConn)
	return nil
}

func (s *Server) OnMessage(conn *conn.Conn) error {
	tlsConn := conn.GetContext().(*tls.Conn)
	if !tlsConn.ConnectionState().HandshakeComplete {
		return tlsConn.Handshake()
	}
	b := make([]byte, 1024)
	n, err := tlsConn.Read(b)
	fmt.Printf("receive %s \n", b[:n])
	if err != nil {
		return err
	}
	tlsConn.Write(b[:n])
	return nil
}

func main() {
	server, err := qio.NewServer(new(Server))
	if err != nil {
		log.Fatal(err)
	}
	server.Serve("tcp", ":9999")
}
