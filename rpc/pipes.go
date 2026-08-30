package rpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Pipes struct{}

type (
	PipeOpenRequest struct {
		Reserved    uint16
		PipeIndex   uint16
		ServiceName string
		Parameter   string
		Version     uint32
	}
	PipeOpenResponse struct {
		PipeIndex       uint16
		Command         uint16
		ServerPipeIndex uint16
		Status          uint16
	}
	PipeDataRequest struct {
		PipeIndex uint16
		Data      []byte
	}
	PipeDataResponse struct{}
	ControlFrame     struct {
		Type    uint8
		Content []byte
	}
)

const CommandPipeOpened = 0x0001

func (c *ControlFrame) UnmarshalBinary(buf []byte) error {
	f := fieldReader{buf: buf}
	c.Type = f.Byte(false)
	c.Content = f.Bytes(len(f.buf))
	return f.Done()
}

func (c *ControlFrame) MarshalBinary() (out []byte, err error) {
	out = append(out, c.Type)
	out = append(out, c.Content...)
	err = nil
	return
}

func (p *PipeOpenRequest) UnmarshalBinary(buf []byte) error {
	f := fieldReader{buf: buf}
	p.Reserved = f.Uint16(false)
	p.PipeIndex = f.Uint16(false)
	p.ServiceName = f.NarrowString()
	p.Parameter = f.NarrowString()
	p.Version = f.Uint32(false)

	return f.Done()
}

const (
	controlFrameConnRequest     = 0x01
	controlFrameTransportParams = 0x03
	controlFrameConnEstablished = 0x04
)

var defaultTransportParameters = []byte{
	0x00, 0x04, 0x00, 0x00, // 1024
	0x00, 0x04, 0x00, 0x00, // 1024
	0x10, 0x00, 0x00, 0x00, // 16
	0x01, 0x00, 0x00, 0x00, // 1
	0x58, 0x02, 0x00, 0x00, // 600
}

func (p *PipeDataRequest) UnmarshalBinary(buf []byte) error {
	f := &fieldReader{buf: buf}
	p.Data = f.Bytes(len(f.buf))
	return f.Done()
}

func (p *PipeOpenResponse) MarshalBinary() (out []byte, err error) {
	out = make([]byte, 8)
	binary.LittleEndian.PutUint16(out, p.PipeIndex)
	binary.LittleEndian.PutUint16(out[2:], p.Command)
	binary.LittleEndian.PutUint16(out[4:], p.ServerPipeIndex)
	binary.LittleEndian.PutUint16(out[6:], p.Status)
	err = nil
	return
}

func (p *Pipes) Open(rq PipeOpenRequest, rs *PipeOpenResponse) error {
	if rq.PipeIndex < 1 || rq.PipeIndex > 15 {
		return fmt.Errorf("bad pipe index: %d", rq.PipeIndex)
	}

	// todo: validate if pipe is already open

	rs.PipeIndex = rq.PipeIndex
	rs.Command = CommandPipeOpened
	rs.ServerPipeIndex = rq.PipeIndex
	rs.Status = 0x0000

	fmt.Printf("PipeOpenResponse: %v\n", rs)

	return nil
}

func (p *Pipes) Data(PipeDataRequest, *PipeDataResponse) error {
	return nil
}

func (p *Pipes) HandleControlFrame(rq ControlFrame, rs *ControlFrame) error {
	switch rq.Type {
	case controlFrameConnRequest:
		if len(rq.Content) < 4 {
			return fmt.Errorf("connection request frame too short (%d)", len(rq.Content))
		}
		rs.Content = bytes.Clone(rq.Content)
		return nil
	case controlFrameConnEstablished:
		if len(rq.Content) > 0 {
			return fmt.Errorf("%d trailing bytes in connection established frame", len(rq.Content))
		}
		rs.Type = controlFrameTransportParams
		rs.Content = bytes.Clone(defaultTransportParameters)
		return nil
	}
	return fmt.Errorf("unknown control frame type 0x%x", rq.Type)
}
