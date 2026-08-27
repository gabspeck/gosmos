package rpc

import (
	"log"
	"net/rpc"
)

func RegisterProcedures(server *rpc.Server) error {
	err := server.Register(&Pipes{})
	if err != nil {
		log.Fatalln(err)
	}
	return err
}
