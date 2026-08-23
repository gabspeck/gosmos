package rpc

type encodable interface {
	appendTo(buf []byte) []byte
	size() uint
}

type StraightRecord struct {
	Length    uint16
	PipeIndex uint8
	PipeMessage
}

func (s *StraightRecord) decodeParams(f *fieldReader) error {
	s.Length = f.Uint16()
	s.PipeIndex = uint8(f.Byte())
	s.PipeMessage.decodeParams(f)

	return f.Done()
}

type PipeMessage interface {
	Routing() uint16
	decodeParams(f *fieldReader) error
}

type ControlFrame struct {
	ControlFrameBody
}

func (cf ControlFrame) Routing() uint16 {
	return 0xFFFF
}

type ControlFrameBody interface {
	Type() uint8
}

type ConnectionRequest struct {
	ControlFrame
	Data []byte
}

func (cr ConnectionRequest) Type() uint8 {
	return 0x01
}

type TransportParameters struct {
	PacketSize uint32
	MaxBytes   uint32
	WindowSize uint32
	AckBehind  uint32
	AckTimeout uint32
	KeepAlive  uint32
}

func (tp TransportParameters) Type() uint8 {
	return 0x03
}

type ConnectionEstablished struct{}

func (ce ConnectionEstablished) Type() uint8 {
	return 0x04
}

type PipeOpenRequest struct {
	Reserved    uint16
	PipeIndex   uint16
	ServiceName string
	Parameter   string
	Version     uint32
}

func (por PipeOpenRequest) Routing() uint16 {
	return 0x0000
}

type PipeOpenResponse struct {
	// does NOT include a routing value
	PipeIndex       uint16
	Command         uint16
	ServerPipeIndex uint16
	Status          uint16
}

type PipeClose struct {
	// implements encodable to append a uint8 with value '1'
	PipeIndex uint16
}

type PipeData struct {
	PipeIndex uint16
	HostBlock
}

type HostBlock struct {
	Class uint8
	HostBlock
	Method    uint8
	RequestID uint32 // VLI
	Body      []encodable
}

type CallBlockRequest struct {
	CallBlock
	// array(s) of send parameters and receive descriptors
}

type TaggedValue interface {
	Tag() uint8
	Value() any
	encodable
}

type TaggedByte struct{}

type ChunkedFieldReference struct {
	StreamID uint8
	Length   uint32
}

type CallBlockResponse struct {
	CallBlock
}
