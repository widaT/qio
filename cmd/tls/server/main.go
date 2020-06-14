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
		fmt.Printf("------ %s", b[:n], err)
		if err != nil {
			return err
		}
		return nil
	}
	server, err := qio.NewServer(handle)
	if err != nil {
		log.Fatal(err)
	}

	server.ServeTLS("tcp", ":9999", "tlsf/cert.pem", "tlsf/key.pem")
}
