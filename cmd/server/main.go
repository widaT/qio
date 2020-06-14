package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/widaT/qio"
)

func main() {

	handle := func(conn net.Conn) error {

		b := make([]byte, 1024)

		n, err := conn.Read(b)

		fmt.Println(n)
		if err != nil {
			//return err
			if err == io.EOF {
				return nil
			}
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
