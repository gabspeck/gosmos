package rpc

import (
	"net"
	"net/rpc"
	"testing"
)

func TestReadRequestHeader(t *testing.T) {
	client, server := net.Pipe()

	codec := NewMosServerCodec(server)

	go client.Write([]byte{0x06, 0x00, 0x00, 0xff, 0xff, 0x04})

	r := rpc.Request{}
	err := codec.ReadRequestHeader(&r)
	if err != nil {
		t.Fatal(err)
	}
	if r.ServiceMethod != "ControlFrame.TransportParameters" {
		t.Fatalf("bad service method: %v", r.ServiceMethod)
	}
	if r.Seq != 0 {
		t.Fatalf("bad seq: %d", r.Seq)
	}
}
