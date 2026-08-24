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
)

func (p *Pipes) Open(PipeOpenRequest, *PipeOpenResponse) error {
	return nil
}
func (p *Pipes) Data() error         {}
func (p *Pipes) ControlFrame() error {}
