package rpc

import (
	"bytes"
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
	ControlFrameRequest struct {
		ControlFrame
	}
	ControlFrameResponse struct {
		ControlFrame
	}
)

func (c *ControlFrame) UnmarshalBinary(buf []byte) error {
	f := fieldReader{buf: buf}
	c.Type = f.Byte()
	c.Content = f.Bytes(len(buf))
	return f.Done()
}

func (c *ControlFrame) MarshalBinary() (out []byte, err error) {
	out = append(out, c.Type)
	out = append(out, c.Content...)
	err = nil
	return
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
}

func (p *PipeDataRequest) UnmarshalBinary(buf []byte) error {
	f := &fieldReader{buf: buf}
	p.PipeIndex = f.Uint16()
	p.Data = f.Bytes(len(f.buf))
	return f.Done()
}

func (c *ControlFrameRequest) UnmarshalBinary(buf []byte) error {
	f := &fieldReader{buf: buf}
	c.Type = f.Byte()
	c.Content = f.Bytes(len(f.buf))
	return f.Done()
}

func (p *Pipes) Open(PipeOpenRequest, *PipeOpenResponse) error {
	return nil
}

func (p *Pipes) Data(PipeDataRequest, *PipeDataResponse) error {
	return nil
}

func (p *Pipes) HandleControlFrame(rq ControlFrameRequest, rs *ControlFrameResponse) error {
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
