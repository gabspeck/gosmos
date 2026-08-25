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

func (p *Pipes) Open(PipeOpenRequest, *PipeOpenResponse) error {
	return nil
}

func (p *Pipes) Data(PipeDataRequest, *PipeDataResponse) error {
	return nil
}

func (p *Pipes) ControlFrame(ControlFrameRequest, *ControlFrameResponse) error {
	return nil
}
