package rpc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"gabriels.io/gosmos/types"
)

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
	Appendable interface {
		Size() int
		Append([]byte) []byte
	}
	PipeFrame struct {
		PipeMessage
	}
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
		HostBlock
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
	methodCall struct {
		class     HostBlockClass
		method    uint8
		requestID uint32
	}

	methodCallRequest struct {
		methodCall
		sendParameters    []types.Arg
		receiveParameters []types.Tag
	}

	GUID             [16]byte
	ControlFrameBody interface {
		Type() ControlFrameType
	}
	ControlFrame struct {
		ControlFrameBody
	}
	ConnectionRequest struct {
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
	KeepAlive           struct{}
	TransportParameters struct {
		packetSize uint32
		maxBytes   uint32
		windowSize uint32
		ackBehind  uint32
		ackTimeout uint32
		keepAlive  *uint32
	}
	ConnectionEstablished struct{}
)

var (
	_ PipeMessage = &PipeOpenRequest{}
	_ PipeMessage = &PipeOpenResponse{}
	_ PipeMessage = &ControlFrame{}
	_ PipeMessage = &PipeData{}

	_ ControlFrameBody = &ConnectionRequest{}
	_ ControlFrameBody = &KeepAlive{}
	_ ControlFrameBody = &ConnectionEstablished{}

	_ HostBlock = CallBlock(nil)
	_ CallBlock = &InterfaceTable{}
	_ CallBlock = &methodCall{}
)

func mustBe[T any](v any) T {
	tv, ok := v.(T)
	if !ok {
		panic(fmt.Sprintf("%v is not %s", reflect.ValueOf(v).Type(), reflect.TypeFor[T]().Name()))
	}
	return tv
}

func (p *PipeFrame) Size() int {
	app := mustBe[Appendable](p.PipeMessage)
	return 2 + app.Size()
}

func (p *PipeFrame) Append(b []byte) []byte {
	app := mustBe[Appendable](p.PipeMessage)
	b = binary.LittleEndian.AppendUint16(b, uint16(p.Routing()))
	return app.Append(b)
}

func (c *ControlFrame) Size() int {
	app := mustBe[Appendable](c.ControlFrameBody)
	return 1 + app.Size()
}

func (c *ControlFrame) Append(b []byte) []byte {
	app := mustBe[Appendable](c.ControlFrameBody)
	b = append(b, byte(c.Type()))
	return app.Append(b)
}

func (t *TransportParameters) Size() int {
	size := 20
	if t.keepAlive != nil {
		size += 4
	}
	return size
}

func (t *TransportParameters) Append(b []byte) []byte {
	b = binary.LittleEndian.AppendUint32(b, t.packetSize)
	b = binary.LittleEndian.AppendUint32(b, t.maxBytes)
	b = binary.LittleEndian.AppendUint32(b, t.windowSize)
	b = binary.LittleEndian.AppendUint32(b, t.ackBehind)
	b = binary.LittleEndian.AppendUint32(b, t.ackTimeout)
	if t.keepAlive != nil {
		b = binary.LittleEndian.AppendUint32(b, *t.keepAlive)
	}
	return b
}

func (c *ConnectionRequest) Size() int {
	return 4 + 4 + len(c.locale) + 1 + len(c.connLog) + 1 + len(c.linkDesc) + 1 + 4 + 28
}

func (c *ConnectionRequest) Append(b []byte) []byte {
	b = binary.LittleEndian.AppendUint32(b, c.formatVer)
	b = binary.LittleEndian.AppendUint32(b, c.lineRate)
	b = append(append(b, c.locale...), 0)
	b = append(append(b, c.connLog...), 0)
	b = append(append(b, c.linkDesc...), 0)
	b = binary.LittleEndian.AppendUint32(b, c.elapsed)
	b = binary.LittleEndian.AppendUint32(b, c.osBlock.languageID)
	b = binary.LittleEndian.AppendUint32(b, c.osBlock.reserved0)
	b = binary.LittleEndian.AppendUint32(b, c.osBlock.platform)
	b = binary.LittleEndian.AppendUint32(b, c.osBlock.major)
	b = binary.LittleEndian.AppendUint32(b, c.osBlock.minor)
	b = binary.LittleEndian.AppendUint32(b, c.osBlock.build)
	b = binary.LittleEndian.AppendUint32(b, c.osBlock.reserved1)
	return b
}

func (m *methodCall) Class() HostBlockClass {
	return m.class
}

func (m *methodCall) Method() uint8 {
	return m.method
}

func (m *methodCall) RequestID() uint32 {
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

func (c *ControlFrame) Routing() RoutingValue {
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

func unmarshalPipeFrame(buf []byte) (*PipeFrame, error) {
	r := fieldReader{buf: buf}
	routing := r.Uint16()
	if r.err != nil {
		return nil, r.err
	}
	var pipeMessage PipeMessage
	var err error
	switch {
	case routing == uint16(routingPipeOpen):
		pipeMessage, err = unmarshalPipeOpenRequest(r)
	case routing <= uint16(routingPipeDataEnd):
		pipeMessage, err = unmarshalPipeCommandOrData(RoutingValue(routing), r)
	case routing == uint16(routingControlFrame):
		pipeMessage, err = unmarshalControlFrame(r)
	default:
		pipeMessage, err = nil, fmt.Errorf("invalid routing value: 0x%2x", routing)
	}
	return &PipeFrame{pipeMessage}, err
}

func unmarshalArg(tag byte, f fieldReader) (types.Arg, error) {
	var parsed types.Arg
	switch tag {
	case byte(types.TagRequestUint8):
		parsed = types.Uint8Arg(f.Byte())
	case byte(types.TagRequestUint16):
		parsed = types.Uint16Arg(f.Uint16())
	case byte(types.TagRequestUint32):
		parsed = types.Uint32Arg(f.Uint32())
	case byte(types.TagRequestVariableLengthField):
		parsed = types.VsizeArg(f.VsizeBytes())
	case byte(types.TagRequestCompressedVariableLengthField):
		parsed = types.VsizeBlocksArg(f.VsizeBlocks())
	default:
		return nil, fmt.Errorf("unknown tag: 0x%02x", tag)
	}

	return parsed, f.err
}

func unmarshalControlFrame(r fieldReader) (*ControlFrame, error) {
	ctlType := r.Byte()
	var body ControlFrameBody
	var err error
	switch ctlType {
	case uint8(cfTypeConnRequest):
		body, err = unmarshalConnectionRequest(r)
	case uint8(cfTypeKeepAlive):
		body, err = unmarshalKeepAlive(r)
	case uint8(cfTypeConnEstablished):
		body, err = unmarshalConnectionEstablished(r)
	default:
		return nil, fmt.Errorf("unknown control frame type: 0x%2x", ctlType)
	}
	return &ControlFrame{body}, err
}

func unmarshalConnectionRequest(r fieldReader) (*ConnectionRequest, error) {
	c := ConnectionRequest{}
	c.formatVer = r.Uint32()
	c.lineRate = r.Uint32()
	c.locale = r.NarrowString()
	c.connLog = r.NarrowString()
	c.linkDesc = r.NarrowString()
	c.elapsed = r.Uint32()
	c.osBlock.languageID = r.Uint32()
	c.osBlock.reserved0 = r.Uint32()
	c.osBlock.platform = r.Uint32()
	c.osBlock.major = r.Uint32()
	c.osBlock.minor = r.Uint32()
	c.osBlock.build = r.Uint32()
	c.osBlock.reserved1 = r.Uint32()

	fmt.Printf("connection request: %+v\n", c)

	return &c, r.Done()
}

func unmarshalKeepAlive(f fieldReader) (*KeepAlive, error) {
	return &KeepAlive{}, f.Done()
}

func unmarshalPipeOpenRequest(f fieldReader) (*PipeOpenRequest, error) {
	p := PipeOpenRequest{}
	p.reserved = f.Uint16()
	p.pipeIndex = f.Uint16()
	p.serviceName = f.NarrowString()
	p.parameter = f.NarrowString()
	p.version = f.Uint32()
	return &p, f.Done()
}

func unmarshalConnectionEstablished(f fieldReader) (*ConnectionEstablished, error) {
	return &ConnectionEstablished{}, f.Done()
}

func unmarshalPipeCommandOrData(pipeIndex RoutingValue, f fieldReader) (PipeMessage, error) {
	/* if the first pipe message byte is 0x01 it may have two meanings:
	a pipe close command if followed by nothing else,
	otherwise it's a host block class ID
	*/
	classOrCmd := f.Byte()
	if classOrCmd == cmdByte && f.Done() == nil {
		p := PipeCloseRequest{}
		p.pipeID = pipeIndex
		return &p, nil
	}
	body, err := unmarshalHostBlock(classOrCmd, f)
	if err != nil {
		return nil, err
	}
	p := PipeData{}
	p.pipeID = pipeIndex
	p.HostBlock = body
	return &p, f.Done()
}

func unmarshalHostBlock(class uint8, f fieldReader) (HostBlock, error) {
	switch {
	case class == uint8(hbClassInterfaceTable):
		return unmarshalInterfaceTable(f)
	case class <= uint8(hbClassCallBlockRecordEnd):
		return unmarshalMethodCallRequest(HostBlockClass(class), f)
	case class == uint8(hbClassStreamFrameMore) || class == uint8(hbClassStreamFrameLast):
		return unmarshalStreamFrame(HostBlockClass(class), f)
	default:
		return nil, fmt.Errorf("unknown host block class: 0x%02x", class)
	}
}

func unmarshalInterfaceTable(f fieldReader) (*InterfaceTable, error) {
	return nil, ErrNotImplemented
}

func unmarshalMethodCallRequest(c HostBlockClass, f fieldReader) (*methodCall, error) {
	m := methodCallRequest{}
	m.class = c
	m.method = f.Byte()
	m.requestID = f.VLI()

	for f.Done() != nil {
		tag := f.Byte()
		if f.err != nil {
			return nil, f.err
		}
		if tag&0b10000000 == 0b10000000 {
			m.receiveParameters = append(m.receiveParameters, types.Tag(tag))
		} else {
			arg, err := unmarshalArg(tag, f)
			if err != nil {
				return nil, err
			}
			m.sendParameters = append(m.sendParameters, arg)
		}
	}

	return nil, ErrNotImplemented
}

func unmarshalStreamFrame(c HostBlockClass, f fieldReader) (HostBlock, error) {
	if c != hbClassStreamFrameMore && c != hbClassStreamFrameLast {
		return nil, fmt.Errorf("host block class is not a stream frame class: 0x%2x", c)
	}
	return nil, ErrNotImplemented
}

func unmarshalIteratorCancel(f fieldReader) (HostBlock, error) {
	return nil, nil
}

func OpenPipe(rq PipeOpenRequest, rs *PipeOpenResponse) error {
	if rq.pipeIndex < 1 || rq.pipeIndex > 15 {
		return fmt.Errorf("bad pipe index: %d", rq.pipeIndex)
	}

	// todo: validate if pipe is already open

	rs.pipeID = RoutingValue(rq.pipeIndex)
	rs.command = 0x00
	rs.serverPipeIndex = rq.pipeIndex
	rs.status = 0x0000

	fmt.Printf("PipeOpenResponse: %v\n", rs)

	return nil
}

func dispatchPipeMessage(_ *Session, p *PipeFrame) (Appendable, error) {
	fmt.Printf("dispatching a %v\n", reflect.TypeOf(p.PipeMessage))
	var pm PipeMessage
	var err error
	switch v := p.PipeMessage.(type) {
	case *ControlFrame:
		pm, err = handleControlFrame(v)
	default:
		pm, err = nil, fmt.Errorf("unhandled pipe message type")
	}
	return &PipeFrame{pm}, err
}

func handleControlFrame(rq *ControlFrame) (*ControlFrame, error) {
	switch rq.ControlFrameBody.(type) {
	case *ConnectionRequest:
		return rq, nil
	case *ConnectionEstablished:
		return &ControlFrame{&defaultTransportParameters}, nil
	}
	return nil, fmt.Errorf("unknown control frame type 0x%x", rq.Type)
}
