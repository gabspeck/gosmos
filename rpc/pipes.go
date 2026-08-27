package rpc

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
	PipeDataResponse    struct{}
	ControlFrameRequest struct {
		Type    uint8
		Content []byte
	}
	ControlFrameResponse struct{}
)

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

func (p *Pipes) ControlFrame(rq ControlFrameRequest, rs *ControlFrameResponse) error {
	switch rq.Type {
	case 0x01:
		return nil
	case 0x03:
	case 0x04:
	}
	return nil
}
