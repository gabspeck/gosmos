package rpc

import (
	"errors"
	"fmt"
)

type Pipes struct{}

var ErrNotImplemented = errors.New("not implemented")

type ControlFrameType uint8

const (
	cfTypeConnRequest     = ControlFrameType(0x01)
	cfTypeKeepAlive       = ControlFrameType(0x02)
	cfTypeTransportParams = ControlFrameType(0x03)
	cfTypeConnEstablished = ControlFrameType(0x04)
)

type RoutingValue uint16

const (
	routingPipeOpen     = RoutingValue(0x0000)
	routingPipeDataEnd  = RoutingValue(0x000F)
	routingControlFrame = RoutingValue(0xFFFF)
)

type HostBlockClass uint8

const (
	hbClassInterfaceTable     = HostBlockClass(0x00)
	hbClassCallBlockRecordEnd = HostBlockClass(0xDF)
	hbClassStreamFrameMore    = HostBlockClass(0xE8)
	hbClassStreamFrameLast    = HostBlockClass(0xE7)
)

const cmdByte = uint8(0x01)

type (
	PipeMessage interface {
		Routing() RoutingValue
	}
	PipeOpenRequest struct {
		reserved    uint16
		pipeIndex   uint16
		serviceName string
		parameter   string
		version     uint32
	}
	pipeBound struct {
		pipeID RoutingValue
	}
	PipeOpenResponse struct {
		pipeBound
		command         uint16
		serverPipeIndex uint16
		status          uint16
	}
	PipeCloseRequest struct {
		pipeBound
	}
	PipeData struct {
		pipeBound
		body HostBlock
	}
	HostBlock interface {
		Class() HostBlockClass
	}
	CallBlock interface {
		HostBlock
		Method() uint8
		RequestID() uint32
	}
	InterfaceTable struct {
		interfaces map[uint8]GUID
	}
	MethodCall struct {
		class     HostBlockClass
		method    uint8
		requestID uint32
	}
	GUID         [16]byte
	ControlFrame interface {
		PipeMessage
		Type() ControlFrameType
	}
	controlFrameMessage struct{}
	ConnectionRequest   struct {
		controlFrameMessage
		formatVer uint32
		lineRate  uint32
		locale    string
		connLog   string
		linkDesc  string
		elapsed   uint32
		osBlock   OSBlock
	}
	OSBlock struct {
		languageID uint32
		reserved0  uint32
		platform   uint32
		major      uint32
		minor      uint32
		build      uint32
		reserved1  uint32
	}
	KeepAlive struct {
		controlFrameMessage
	}
	TransportParameters struct {
		controlFrameMessage
		packetSize uint32
		maxBytes   uint32
		windowSize uint32
		ackBehind  uint32
		ackTimeout uint32
		keepAlive  *uint32
	}
	ConnectionEstablished struct {
		controlFrameMessage
	}
)

var (
	_ PipeMessage = &PipeOpenRequest{}
	_ PipeMessage = &PipeOpenResponse{}
	_ PipeMessage = ControlFrame(nil)
	_ PipeMessage = &controlFrameMessage{}
	_ PipeMessage = &PipeData{}

	_ ControlFrame = &ConnectionRequest{}
	_ ControlFrame = &KeepAlive{}
	_ ControlFrame = &TransportParameters{}
	_ ControlFrame = &ConnectionEstablished{}

	_ HostBlock = CallBlock(nil)
	_ CallBlock = &InterfaceTable{}
	_ CallBlock = &MethodCall{}
)

func (m *MethodCall) Class() HostBlockClass {
	return m.class
}

func (m *MethodCall) Method() uint8 {
	return m.method
}

func (m *MethodCall) RequestID() uint32 {
	return m.requestID
}

func (c *ConnectionEstablished) Type() ControlFrameType {
	return 0x04
}

func (t *TransportParameters) Type() ControlFrameType {
	return 0x03
}

func (k *KeepAlive) Type() ControlFrameType {
	return 0x02
}

func (c *ConnectionRequest) Type() ControlFrameType {
	return cfTypeConnRequest
}

func (c *controlFrameMessage) Routing() RoutingValue {
	return 0xFFFF
}

func (p *PipeOpenRequest) Routing() RoutingValue {
	return 0x0000
}

func (p *pipeBound) Routing() RoutingValue {
	return p.pipeID
}

func (i *InterfaceTable) Class() HostBlockClass {
	return 0x00
}

func (i *InterfaceTable) Method() uint8 {
	return 0x00
}

func (i *InterfaceTable) RequestID() uint32 {
	return 0x00
}

func (g GUID) String() string {
	// todo: proper guid formatting
	return fmt.Sprintf("%d", g)
}

func unmarshalPipeMessage(r fieldReader) (PipeMessage, error) {
	routing := r.Uint16(false)
	if r.err != nil {
		return nil, r.err
	}
	switch {
	case routing == uint16(routingPipeOpen):
		return decodePipeOpenRequest(r)
	case routing <= uint16(routingPipeDataEnd):
		return decodePipeCommandOrData(RoutingValue(routing), r)
	case routing == uint16(routingControlFrame):
		return decodeControlFrame(r)
	default:
		return nil, fmt.Errorf("invalid routing value: 0x%2x", routing)
	}
}

func decodeControlFrame(r fieldReader) (ControlFrame, error) {
	ctlType := r.Byte(false)
	switch ctlType {
	case uint8(cfTypeConnRequest):
		return decodeConnectionRequest(r)
	case uint8(cfTypeKeepAlive):
		return decodeKeepAlive(r)
	case uint8(cfTypeTransportParams):
		return decodeTransportParameters(r)
	case uint8(cfTypeConnEstablished):
		return decodeConnectionEstablished(r)
	}
	return nil, fmt.Errorf("unknown control frame type: 0x%2x", ctlType)
}

func decodeConnectionRequest(r fieldReader) (*ConnectionRequest, error) {
	c := ConnectionRequest{}
	c.formatVer = r.Uint32(false)
	c.lineRate = r.Uint32(false)
	c.locale = r.NarrowString()
	c.connLog = r.NarrowString()
	c.linkDesc = r.NarrowString()
	c.elapsed = r.Uint32(false)
	c.osBlock.languageID = r.Uint32(false)
	c.osBlock.reserved0 = r.Uint32(false)
	c.osBlock.platform = r.Uint32(false)
	c.osBlock.major = r.Uint32(false)
	c.osBlock.minor = r.Uint32(false)
	c.osBlock.build = r.Uint32(false)
	c.osBlock.reserved1 = r.Uint32(false)

	return &c, r.Done()
}

func decodeKeepAlive(f fieldReader) (*KeepAlive, error) {
	return &KeepAlive{}, f.Done()
}

func decodeTransportParameters(f fieldReader) (*TransportParameters, error) {
	// server only
	return nil, ErrNotImplemented
}

func decodePipeOpenRequest(f fieldReader) (*PipeOpenRequest, error) {
	p := PipeOpenRequest{}
	p.reserved = f.Uint16(false)
	p.pipeIndex = f.Uint16(false)
	p.serviceName = f.NarrowString()
	p.parameter = f.NarrowString()
	p.version = f.Uint32(false)
	return &p, f.Done()
}

func decodeConnectionEstablished(f fieldReader) (*ConnectionEstablished, error) {
	return &ConnectionEstablished{}, f.Done()
}

func decodePipeCommandOrData(pipeIndex RoutingValue, f fieldReader) (PipeMessage, error) {
	/* if the first pipe message byte is 0x01 it may have two meanings:
	a pipe close command if followed by nothing else,
	otherwise it's a host block class ID
	*/
	classOrCmd := f.Byte(false)
	if classOrCmd == cmdByte && f.Done() == nil {
		p := PipeCloseRequest{}
		p.pipeID = pipeIndex
		return &p, nil
	}
	body, err := decodeHostBlock(classOrCmd, f)
	if err != nil {
		return nil, err
	}
	p := PipeData{}
	p.pipeID = pipeIndex
	p.body = body
	return &p, f.Done()
}

func decodeHostBlock(class uint8, f fieldReader) (HostBlock, error) {
	switch {
	case class == uint8(hbClassInterfaceTable):
		return decodeInterfaceTable(f)
	case class <= uint8(hbClassCallBlockRecordEnd):
		return decodeMethodCall(HostBlockClass(class), f)
	case class == uint8(hbClassStreamFrameMore) || class == uint8(hbClassStreamFrameLast):
		return decodeStreamFrame(HostBlockClass(class), f)
	default:
		return nil, fmt.Errorf("unknown host block class: 0x%02x", class)
	}
}

func decodeInterfaceTable(f fieldReader) (*InterfaceTable, error) {
	return nil, ErrNotImplemented
}

func decodeMethodCall(c HostBlockClass, f fieldReader) (*MethodCall, error) {
	m := MethodCall{}
	m.class = c
	m.method = f.Byte(false)
	m.requestID = f.VLI()

	return nil, ErrNotImplemented
}

func decodeStreamFrame(c HostBlockClass, f fieldReader) (HostBlock, error) {
	if c != hbClassStreamFrameMore && c != hbClassStreamFrameLast {
		return nil, fmt.Errorf("host block class is not a stream frame class: 0x%2x", c)
	}
	return nil, ErrNotImplemented
}

func decodeIteratorCancel(f fieldReader) (HostBlock, error) {
	return nil, nil
}

// const CommandPipeOpened = 0x0001

// func (c *controlFrameMessage) UnmarshalBinary(buf []byte) error {
// 	f := fieldReader{buf: buf}
// 	c.Type = f.Byte(false)
// 	c.Content = f.Bytes(len(f.buf))
// 	return f.Done()
// }

// func (c *ControlFrame) MarshalBinary() (out []byte, err error) {
// 	out = append(out, c.Type)
// 	out = append(out, c.Content...)
// 	err = nil
// 	return
// }

// func (p *PipeOpenRequest) UnmarshalBinary(buf []byte) error {
// 	f := fieldReader{buf: buf}
// 	p.Reserved = f.Uint16(false)
// 	p.PipeIndex = f.Uint16(false)
// 	p.ServiceName = f.NarrowString()
// 	p.Parameter = f.NarrowString()
// 	p.Version = f.Uint32(false)

// 	return f.Done()
// }

// const (
// 	controlFrameConnRequest     = 0x01
// 	controlFrameTransportParams = 0x03
// 	controlFrameConnEstablished = 0x04
// )

// var defaultTransportParameters = []byte{
// 	0x00, 0x04, 0x00, 0x00, // 1024
// 	0x00, 0x04, 0x00, 0x00, // 1024
// 	0x10, 0x00, 0x00, 0x00, // 16
// 	0x01, 0x00, 0x00, 0x00, // 1
// 	0x58, 0x02, 0x00, 0x00, // 600
// }

// func (p *PipeDataRequest) UnmarshalBinary(buf []byte) error {
// 	f := &fieldReader{buf: buf}
// 	p.Data = f.Bytes(len(f.buf))
// 	return f.Done()
// }

// func (p *PipeOpenResponse) MarshalBinary() (out []byte, err error) {
// 	out = make([]byte, 8)
// 	binary.LittleEndian.PutUint16(out, p.PipeIndex)
// 	binary.LittleEndian.PutUint16(out[2:], p.Command)
// 	binary.LittleEndian.PutUint16(out[4:], p.ServerPipeIndex)
// 	binary.LittleEndian.PutUint16(out[6:], p.Status)
// 	err = nil
// 	return
// }

// func (p *Pipes) Open(rq PipeOpenRequest, rs *PipeOpenResponse) error {
// 	if rq.PipeIndex < 1 || rq.PipeIndex > 15 {
// 		return fmt.Errorf("bad pipe index: %d", rq.PipeIndex)
// 	}

// 	// todo: validate if pipe is already open

// 	rs.PipeIndex = rq.PipeIndex
// 	rs.Command = CommandPipeOpened
// 	rs.ServerPipeIndex = rq.PipeIndex
// 	rs.Status = 0x0000

// 	fmt.Printf("PipeOpenResponse: %v\n", rs)

// 	return nil
// }

// func (p *Pipes) Data(PipeDataRequest, *PipeDataResponse) error {
// 	return nil
// }

// func (p *Pipes) HandleControlFrame(rq ControlFrame, rs *ControlFrame) error {
// 	switch rq.Type {
// 	case controlFrameConnRequest:
// 		if len(rq.Content) < 4 {
// 			return fmt.Errorf("connection request frame too short (%d)", len(rq.Content))
// 		}
// 		rs.Content = bytes.Clone(rq.Content)
// 		return nil
// 	case controlFrameConnEstablished:
// 		if len(rq.Content) > 0 {
// 			return fmt.Errorf("%d trailing bytes in connection established frame", len(rq.Content))
// 		}
// 		rs.Type = controlFrameTransportParams
// 		rs.Content = bytes.Clone(defaultTransportParameters)
// 		return nil
// 	}
// 	return fmt.Errorf("unknown control frame type 0x%x", rq.Type)
// }
