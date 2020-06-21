package main

import (
	"fmt"
	"io"
	"log"

	"github.com/widaT/qio"
)

type Server struct {
	*qio.DefaultEvServer
}

func (s *Server) OnMessage(conn *qio.Conn) error {
	b := make([]byte, 0x10000)
	n, err := conn.Read(b)
	if err != nil {
		//return err
		if err == io.EOF {
			return nil
		}
		return err
	}
	fmt.Printf("%s", b[:n])
	return nil
}

func main() {
	server, err := qio.NewServer(new(Server))
	if err != nil {
		log.Fatal(err)
	}
	server.Serve("tcp", ":9999")
}
