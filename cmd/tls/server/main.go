package main

import (
	"fmt"
	"log"

	"github.com/widaT/qio"
	tls "github.com/widaT/tls13"
)

type Server struct {
	*qio.DefaultEvServer
}

var tlsConfig *tls.Config

func init() {
	c, err := tls.LoadX509KeyPair("tls/cert.pem", "tls/key.pem")
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
	if !tlsConn.ConnectionState().HandshakeComplete {
		err := tlsConn.Handshake()
		if err != nil {
			if tls.StatusPartial == err {
				return nil
			}
			return err
		}
		return nil
	}
	b := make([]byte, 1024)
	n, err := tlsConn.Read(b)
	if tls.StatusPartial == err {
		return nil
	}
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
