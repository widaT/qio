package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"
)

func main() {
	tks()
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
	buf := make([]byte, 100)
	for {
		time.Sleep(1e9)

		n, err := conn.Write([]byte("aaaaaaaa"))
		if err != nil {
			log.Println(n, err)
			return
		}
		n, err = conn.Read(buf)
		if err != nil {
			log.Println(n, err)
			return
		}
		fmt.Printf("%s\n", buf[:n])
	}
}
