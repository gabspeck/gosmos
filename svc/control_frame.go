package rpc

import "fmt"

type ConnectionEstablishedArgs struct{}

func (*ControlFrame) ConnectionEstablished(args *ConnectionEstablishedArgs, tp *TransportParameters) error {
	fmt.Printf("cea -> %v\n", args)
	*tp = TransportParameters{}
	return nil
}
