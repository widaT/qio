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
		fmt.Printf("%s", b[:n])
		if err != nil {
			fmt.Println("err-----", err)
			return err
		}
		return nil
	}
	server, err := qio.NewServer(handle)
	if err != nil {
		log.Fatal(err)
	}

	server.ServeTLs("tcp", ":9999", "tls/cert.pem", "tls/key.pem")
}
