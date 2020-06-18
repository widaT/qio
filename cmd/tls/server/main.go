package main

import (
	"fmt"
	"log"

	"github.com/widaT/qio"
	"github.com/widaT/qio/conn"
)

type Server struct {
	*qio.DefaultEvServer
}

func (s *Server) OnMessage(conn conn.Conn) error {
	b := make([]byte, 1024)
	n, err := conn.Read(b)
	fmt.Printf("receive %s \n", b[:n])
	if err != nil {
		return err
	}

	conn.Write(b[:n])

	return nil
}

func main() {
	server, err := qio.NewServer(new(Server))
	if err != nil {
		log.Fatal(err)
	}

	server.ServeTLS("tcp", ":9999", "tlsf/cert.pem", "tlsf/key.pem")
}
