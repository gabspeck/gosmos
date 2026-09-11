package types

type (
	Arg interface {
		Tag() Tag
	}
	Tag            uint8
	Uint8Arg       uint8
	Uint16Arg      uint16
	Uint32Arg      uint32
	VsizeArg       []byte
	VsizeBlocksArg [][]byte
)

const (
	TagRequestUint8                         = Tag(0x01)
	TagRequestUint16                        = Tag(0x02)
	TagRequestUint32                        = Tag(0x03)
	TagRequestVariableLengthField           = Tag(0x04)
	TagRequestChunkedRef                    = Tag(0x05)
	TagRequestCompressedVariableLengthField = Tag(0x44)
	TagRequestCompressedChunkedRef          = Tag(0x45)
)

func (u Uint8Arg) Tag() Tag {
	return TagRequestUint8
}

func (u Uint16Arg) Tag() Tag {
	return TagRequestUint16
}

func (u Uint32Arg) Tag() Tag {
	return TagRequestUint32
}

func (v VsizeArg) Tag() Tag {
	return TagRequestVariableLengthField
}

func (v VsizeBlocksArg) Tag() Tag {
	return TagRequestCompressedVariableLengthField
}
