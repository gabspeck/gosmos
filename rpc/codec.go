package rpc

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"net/rpc"
)

var _ rpc.ServerCodec = &MosServerCodec{}

type MosServerCodec struct {
	seq  uint64
	len  int
	conn io.ReadWriteCloser
}

func NewMosServerCodec(conn io.ReadWriteCloser) *MosServerCodec {
	return &MosServerCodec{
		conn: conn,
	}
}

// Close implements [rpc.ServerCodec].
func (m *MosServerCodec) Close() error {
	panic("unimplemented")
}

// ReadRequestBody implements [rpc.ServerCodec].
func (m *MosServerCodec) ReadRequestBody(p any) error {
	bu, ok := p.(encoding.BinaryUnmarshaler)
	if !ok {
		return fmt.Errorf("not an unmarshaler")
	}

	buf := make([]byte, m.len)

	if _, err := io.ReadFull(m.conn, buf); err != nil {
		return err
	}

	if err := bu.UnmarshalBinary(buf); err != nil {
		return err
	}

	return nil
}

// ReadRequestHeader implements [rpc.ServerCodec].
func (m *MosServerCodec) ReadRequestHeader(r *rpc.Request) error {
	r.Seq = m.seq
	m.seq += 1

	var header [5]byte

	if _, err := io.ReadFull(m.conn, header[:]); err != nil {
		return err
	}

	size := int(binary.LittleEndian.Uint16(header[:]))
	if size < len(header) {
		return fmt.Errorf("bad packet length: %d", size)
	}

	// cmd := uint8(header[2])

	switch routing := binary.LittleEndian.Uint16(header[3:]); {
	case routing == 0x0000:
		r.ServiceMethod = "Pipes.Open"
	case routing >= 0x0001 && routing <= 0x000F:
		r.ServiceMethod = "Pipes.Data"
	case routing == 0xFFFF:
		r.ServiceMethod = "Pipes.ControlFrame"
	default:
		return fmt.Errorf("unhandled routing: 0x%02x", routing)
	}

	m.len = size - len(header)
	return nil
}

// WriteResponse implements [rpc.ServerCodec].
func (m *MosServerCodec) WriteResponse(r *rpc.Response, p any) error {
	panic("unimplemented")
}
