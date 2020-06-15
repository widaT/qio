package main

import (
	"crypto/tls"
	"log"
	"net"
	"time"
)

func tcp() {

	conn, err := net.Dial("tcp", ":9999")
	if err != nil {
		log.Fatal(err)
	}

	b := make([]byte, 12087)
	for {
		time.Sleep(1e9)

		n, err := conn.Write(b)
		if err != nil {
			log.Println(n, err)
			return
		}
	}

}

func main() {

	tcp()
}

func tks() {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", "localhost:9999", conf)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	b := make([]byte, 12087)
	for {
		time.Sleep(1e9)

		n, err := conn.Write(b)
		if err != nil {
			log.Println(n, err)
			return
		}
	}

	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(n, err)
		return
	}
	println(string(buf[:n]))
}
