package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"gabriels.io/gosmos/rpc"
)

func main() {
	l, err := net.Listen("tcp", ":5690")
	if err != nil {
		log.Fatalln(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // unregisters the signal handler so a second ctrl+c force quits
	fmt.Println("starting server")
	if err := rpc.NewServer().Serve(ctx, l); err != nil {
		log.Fatal(err)
	}
	// 1a0000ffff030004000000040000100000000100000058020000
}
