package main

import (
	"fmt"
	"log"
	"net"

	"github.com/widaT/qio"
)

func main() {

	handle := func(conn net.Conn) error {
		b := make([]byte, 1024)

		n, err := conn.Read(b)

		if err != nil {
			return err
		}
		fmt.Printf("------%s", b[:n])
		return nil
	}
	server, err := qio.NewServer(handle)
	if err != nil {
		log.Fatal(err)
	}

	server.Serve("tcp", ":9999")

}
