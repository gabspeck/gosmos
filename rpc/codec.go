package rpc

import (
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"sync"
)

var _ rpc.ServerCodec = &MosServerCodec{}

type MosServerCodec struct {
	seq     uint64
	len     int
	conn    io.ReadWriteCloser
	lock    sync.Mutex
	pending pendingRequestMap
}

type pendingRequestMap map[uint64]requestContext

type requestContext struct {
	routing uint16
	cmd     uint8
}

func NewMosServerCodec(conn io.ReadWriteCloser) *MosServerCodec {
	return &MosServerCodec{
		conn:    conn,
		lock:    sync.Mutex{},
		pending: make(pendingRequestMap),
	}
}

// Close implements [rpc.ServerCodec].
func (m *MosServerCodec) Close() error {
	return m.conn.Close()
}

// ReadRequestBody implements [rpc.ServerCodec].
func (m *MosServerCodec) ReadRequestBody(p any) error {
	if p == nil {
		fmt.Printf("[%d] nil request body received\n", m.seq)
		// todo: actually handle this
		delete(m.pending, m.seq)
		return m.Close()
	}
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

	cmd := uint8(header[2])
	routing := binary.LittleEndian.Uint16(header[3:])
	fmt.Printf("-> [%d] routing: 0x%04x\n", r.Seq, routing)
	switch {
	case routing == 0x0000:
		r.ServiceMethod = "Pipes.Open"
	case routing >= 0x0001 && routing <= 0x000F:
		r.ServiceMethod = "Pipes.Data"
	case routing == 0xFFFF:
		r.ServiceMethod = "Pipes.HandleControlFrame"
	default:
		err := fmt.Errorf("unhandled routing: 0x%02x", routing)
		log.Println(err)
		return err
	}
	fmt.Printf("-> [%d] method: %s\n", r.Seq, r.ServiceMethod)
	m.len = size - len(header)
	m.lock.Lock()
	m.pending[r.Seq] = requestContext{
		routing: routing,
		cmd:     cmd,
	}
	m.lock.Unlock()
	return nil
}

// WriteResponse implements [rpc.ServerCodec].
func (m *MosServerCodec) WriteResponse(r *rpc.Response, p any) error {
	context, ok := m.pending[r.Seq]
	if r.Error != "" {
		log.Println(r.Error)
		return errors.New(r.Error)
	}
	if !ok {
		return fmt.Errorf("context not found for request %d", r.Seq)
	}

	bm, ok := p.(encoding.BinaryMarshaler)
	if !ok {
		err := fmt.Errorf("don't know how to marshal response from method %s", r.ServiceMethod)
		log.Println(err)
		return err
	}

	payload := make([]byte, 4)

	m.lock.Lock()
	binary.LittleEndian.AppendUint16(payload, context.routing)
	delete(m.pending, m.seq)
	m.lock.Unlock()
	response, err := bm.MarshalBinary()
	if err != nil {
		return err
	}
	payload = append(payload, response...)
	fmt.Printf("<- [%d/%s]: 0x%x\n", r.Seq, r.ServiceMethod, payload)
	_, err = m.conn.Write(payload)
	if err != nil {
		return err
	}

	return nil
}
