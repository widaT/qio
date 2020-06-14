package main

import (
	"crypto/tls"
	"log"
	"time"
)

func main() {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", "localhost:9999", conf)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		time.Sleep(1e9)
		n, err := conn.Write([]byte("hello\n"))
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
