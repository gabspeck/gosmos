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

func (f *fieldReader) fixed(isTagged bool, want Tag, size int) []byte {
	if f.err != nil {
		return nil
	}
	tagSize := 0
	if isTagged {
		tagSize = 1
	}
	if len(f.buf) < tagSize+size {
		f.err = fmt.Errorf("field %d: truncated", f.fieldIndex)
		return nil
	}
	if isTagged && f.buf[0] != uint8(want) {
		f.err = fmt.Errorf("param %d: want tag 0x%02x, got 0x%02x", f.fieldIndex, want, f.buf[0])
		return nil
	}
	v := f.buf[tagSize : tagSize+size]
	f.buf, f.fieldIndex = f.buf[tagSize+size:], f.fieldIndex+1
	return v
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
	f.buf, f.fieldIndex = f.buf[termIdx+1:], f.fieldIndex+1
	return v
}

func (f *fieldReader) Byte(tagged bool) uint8 {
	b := f.fixed(tagged, TagRequestUint8, 1)
	if f.err != nil {
		return 0
	}
	return b[0]
}

func (f *fieldReader) Uint16(tagged bool) uint16 {
	b := f.fixed(tagged, TagRequestUint16, 2)
	if f.err != nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

func (f *fieldReader) Uint32(tagged bool) uint32 {
	b := f.fixed(tagged, TagRequestUint32, 4)
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

func (f *fieldReader) Bytes(n int) []byte {
	if f.err != nil {
		return nil
	}
	if n > len(f.buf) {
		f.err = fmt.Errorf("wanted %d bytes, only %d available", n, len(f.buf))
		return nil
	}
	v := f.buf[:n]
	f.buf, f.fieldIndex = f.buf[n:], f.fieldIndex+1
	return v
}

func (f *fieldReader) Done() error {
	if f.err != nil {
		return f.err
	}
	if len(f.buf) > 0 {
		return fmt.Errorf("%d unexpected trailing bytes", len(f.buf))
	}
	return nil
}
