package rpc

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/rpc"
)

var _ rpc.ServerCodec = &MosStraightServerCodec{}

type MosStraightServerCodec struct {
	seq  uint64
	conn io.ReadWriteCloser
}

func NewMosStraightServerCodec(conn io.ReadWriteCloser) *MosStraightServerCodec {
	return &MosStraightServerCodec{
		conn: conn,
	}
}

// Close implements [rpc.ServerCodec].
func (m *MosStraightServerCodec) Close() error {
	panic("unimplemented")
}

// ReadRequestBody implements [rpc.ServerCodec].
func (m *MosStraightServerCodec) ReadRequestBody(any) error {
	panic("unimplemented")
}

// ReadRequestHeader implements [rpc.ServerCodec].
func (m *MosStraightServerCodec) ReadRequestHeader(r *rpc.Request) error {
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

	return nil
}

// WriteResponse implements [rpc.ServerCodec].
func (m *MosStraightServerCodec) WriteResponse(*rpc.Response, any) error {
	panic("unimplemented")
}
