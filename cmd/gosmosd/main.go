package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	gosmosrpc "gabriels.io/gosmos/rpc"
)

func main() {
	l, err := net.Listen("tcp", ":5690")
	if err != nil {
		log.Fatalln(err)
	}
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := l.Accept()
		fmt.Printf("new connection from %s\n", conn.RemoteAddr())
		if err != nil {
			log.Println(err)
			continue
		}
		go rpc.ServeCodec(gosmosrpc.NewMosServerCodec(conn))
	}
}
