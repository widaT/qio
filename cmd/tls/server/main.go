package main

import (
	"fmt"
	"log"
	"net"

	"github.com/widaT/qio"
)

func main() {

	handle := func(conn net.Conn) error {
		b := make([]byte, 0x10000)
		n, err := conn.Read(b)
		fmt.Printf("receive %d \n", n)
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
