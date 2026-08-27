package rpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/rpc"
	"testing"
	"time"
)

func TestReadRequestHeader(t *testing.T) {
	testCases := []struct {
		description string
		method      string
		error       string
		length      [2]byte
		routing     [2]byte
	}{
		{
			description: "bad length",
			error:       "bad packet length: 4",
			length:      [2]byte{0x04, 0x00},
			routing:     [2]byte{0x00, 0x00},
		},
		{
			description: "longer length",
			method:      "Pipes.Data",
			length:      [2]byte{0xff, 0xff},
			routing:     [2]byte{0x02, 0x00},
		},
		{
			description: "unknown routing",
			error:       "unhandled routing: 0xcac0",
			length:      [2]byte{0x05, 0x00},
			routing:     [2]byte{0xc0, 0xca},
		},
		{
			method:  "Pipes.Open",
			length:  [2]byte{0x05, 0x00},
			routing: [2]byte{0x00, 0x00},
		},
		{
			method:  "Pipes.Data",
			length:  [2]byte{0x05, 0x00},
			routing: [2]byte{0x01, 0x00},
		},
		{
			method:  "Pipes.ControlFrame",
			length:  [2]byte{0x05, 0x00},
			routing: [2]byte{0xff, 0xff},
		},
	}
	for _, tC := range testCases {
		if tC.description == "" {
			tC.description = tC.method
		}
		t.Run(tC.description, func(t *testing.T) {
			server, client := net.Pipe()
			defer func() {
				server.Close()
				client.Close()
			}()
			server.SetDeadline(time.Now().Add(time.Second))
			client.SetDeadline(time.Now().Add(time.Second))

			result := make(chan error, 1)

			codec := NewMosServerCodec(server)

			req := rpc.Request{}
			go client.Write([]byte{tC.length[0], tC.length[1], 0x00, tC.routing[0], tC.routing[1]})
			go func() {
				result <- codec.ReadRequestHeader(&req)
			}()

			err := <-result
			if (err == nil) != (tC.error == "") {
				t.Fatalf("wanted error %q, got %v", tC.error, err)
			}
			if err != nil && err.Error() != tC.error {
				t.Fatalf("wanted error %q, got %v", tC.error, err)
			}

			if req.Seq != 0 {
				t.Fatalf("wanted req.Seq %d; got %d", 0, req.Seq)
			}

			if req.ServiceMethod != tC.method {
				t.Fatalf("wanted req.ServiceMethod %q; got %q", tC.method, req.ServiceMethod)
			}
		})
	}
}

type TestProcedureParams struct {
	expectedBuf []byte
	SomeUint16  uint16
}

func (t *TestProcedureParams) UnmarshalBinary(buf []byte) error {
	if !bytes.Equal(t.expectedBuf, buf) {
		return fmt.Errorf("expected to receive buffer 0x%x; got 0x%x", t.expectedBuf, buf)
	}
	t.SomeUint16 = binary.LittleEndian.Uint16(buf[1:])
	return nil
}

func TestReadRequestBody(t *testing.T) {
	server, client := net.Pipe()
	defer func() {
		server.Close()
		client.Close()
	}()
	server.SetDeadline(time.Now().Add(time.Second))
	client.SetDeadline(time.Now().Add(time.Second))

	result := make(chan error, 1)

	buf := []byte{TagRequestUint16, 0xca, 0xc0}

	go client.Write(buf)

	codec := &MosServerCodec{conn: server, len: len(buf)}
	params := TestProcedureParams{expectedBuf: buf}
	go func() {
		result <- codec.ReadRequestBody(&params)
	}()

	if err := <-result; err != nil {
		t.Fatal(err)
	}
	const want = 0xc0ca
	if params.SomeUint16 != want {
		t.Fatalf("wanted 0x%04x; got 0x%04x", want, params.SomeUint16)
	}
}
