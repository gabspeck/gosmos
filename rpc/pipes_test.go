package rpc

import (
	"net"
	"testing"
	"time"
)

func TestHandleConnectionEstablished(t *testing.T) {
	client, server := net.Pipe()
	client.SetDeadline(time.Now().Add(time.Second))
	server.SetDeadline(time.Now().Add(time.Second))
}
