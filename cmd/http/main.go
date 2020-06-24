package main

import (
	"fmt"
	"log"

	http "github.com/widaT/http1"
	"github.com/widaT/qio"
)

type Server struct {
	*qio.DefaultEvServer
}

func (s *Server) OnMessage(conn *qio.Conn) error {
	req := new(http.Request)

	err := req.Parse(conn)
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	//fmt.Println(req)

	fmt.Println(*req.Header())
	fmt.Println(req.Body())

	return nil
}

func main() {
	server, err := qio.NewServer(new(Server))
	if err != nil {
		log.Fatal(err)
	}
	server.Serve("tcp", ":9999")
}
