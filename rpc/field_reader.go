package rpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type fieldReader struct {
	buf        []byte
	fieldIndex int
	err        error
}

type Tag uint8

const (
	TagRequestUint8                         = 0x01
	TagRequestUint16                        = 0x02
	TagRequestUint32                        = 0x03
	TagRequestVariableLengthField           = 0x04
	TagRequestChunkedRef                    = 0x05
	TagRequestCompressedVariableLengthField = 0x44
	TagRequestCompressedChunkedRef          = 0x45
)

func (f *fieldReader) tagged(want Tag, size int) []byte {
	if f.err != nil {
		return nil
	}
	if len(f.buf) < 1+size {
		f.err = fmt.Errorf("field %d: truncated", f.fieldIndex)
		return nil
	}
	if f.buf[0] != uint8(want) {
		f.err = fmt.Errorf("param %d: want tag 0x%02x, got 0x%02x", f.fieldIndex, want, f.buf[0])
		return nil
	}
	v := f.buf[1 : 1+size]
	f.buf, f.fieldIndex = f.buf[1+size:], f.fieldIndex+1
	return v
}

func (f *fieldReader) Byte() uint8 {
	b := f.tagged(TagRequestUint8, 1)
	if f.err != nil {
		return 0
	}
	return b[0]
}

func (f *fieldReader) terminated(term byte) []byte {
	if f.err != nil {
		return nil
	}
	termIdx := bytes.IndexByte(f.buf, term)
	if termIdx < 0 {
		f.err = fmt.Errorf("field %d: terminator 0x%02x not found", f.fieldIndex, term)
		return nil
	}
	v := f.buf[:termIdx] // skip terminator
	f.buf, f.fieldIndex = f.buf[termIdx:], f.fieldIndex+1
	return v
}

func (f *fieldReader) Uint16() uint16 {
	b := f.tagged(TagRequestUint16, 2)
	if f.err != nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

func (f *fieldReader) Uint32() uint32 {
	b := f.tagged(TagRequestUint32, 4)
	if f.err != nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}

func (f *fieldReader) NarrowString() string {
	b := f.terminated(byte(0))
	if f.err != nil {
		return ""
	}
	return string(b)
}

func (f *fieldReader) Done() error {
	if f.err != nil {
		return f.err
	}
	if len(f.buf) > 0 {
		return fmt.Errorf("unexpected trailing fields")
	}
	return nil
}
