package main

import (
	"fmt"
	"log"
	"net"

	"gabriels.io/gosmos/mos"
)

func main() {
	l, err := net.Listen("tcp", ":5690")
	if err != nil {
		log.Fatalln(err)
	}
	server := mos.NewServer()
	for {
		conn, err := l.Accept()
		fmt.Printf("new connection from %s\n", conn.RemoteAddr())
		if err != nil {
			log.Println(err)
			continue
		}
		go server.HandleConnection(conn)
	}
}
