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

func (f *fieldReader) fixed(size int, complete bool) []byte {
	if f.err != nil {
		return nil
	}
	if len(f.buf) < size {
		f.err = fmt.Errorf("field %d: truncated", f.fieldIndex)
		return nil
	}
	v := f.buf[:size]
	f.buf = f.buf[size:]
	if complete {
		f.fieldIndex += 1
	}
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

func (f *fieldReader) Byte() uint8 {
	b := f.fixed(1, true)
	if f.err != nil {
		return 0
	}
	return b[0]
}

func (f *fieldReader) Uint16() uint16 {
	b := f.fixed(2, true)
	if f.err != nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

func (f *fieldReader) Uint32() uint32 {
	b := f.fixed(4, true)
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

func (f *fieldReader) VLI() uint32 {
	const (
		oneByte   = 0b00000000
		twoBytes  = 0b10000000
		fourBytes = 0b11000000
	)
	buf := f.fixed(1, false)
	if f.err != nil {
		return 0
	}
	first := buf[0]
	formBits := first & 0b11000000

	result := uint32(first & 0b00111111)

	switch formBits {
	case oneByte:
		return result
	case twoBytes:
		buf = f.fixed(1, false)
		if f.err != nil {
			return 0
		}
		second := buf[0]
		return result<<8 | uint32(second)
	case fourBytes:
		buf := make([]byte, 1, 4)
		buf[0] = first
		buf = append(buf, f.fixed(3, false)...)
		if f.err != nil {
			return 0
		}
		f.fieldIndex += 1
		return binary.BigEndian.Uint32(buf) & 0x3FFFFFFF
	}

	f.err = fmt.Errorf("invalid VLI form bits: 0b%s", fmt.Sprintf("%08b", formBits)[0:2])
	return 0
}

func (f *fieldReader) vsize() uint32 {
	buf := f.fixed(1, false)
	if f.err != nil {
		return 0
	}

	first := uint32(buf[0])

	if first&0b1000_0000 == 0b1000_0000 {
		return first & 0b0111_1111
	}

	buf = f.fixed(1, false)
	if f.err != nil {
		return 0
	}

	second := uint32(buf[0])
	if f.err != nil {
		return 0
	}

	f.fieldIndex += 1

	return first<<8 | second
}

func (f *fieldReader) VsizeBytes() []byte {
	size := int(f.vsize())

	return f.fixed(size, true)
}

func (f *fieldReader) VsizeBlocks() [][]byte {
	totalLength := int(f.vsize())

	if f.err != nil {
		return nil
	}

	buf := make([][]byte, 0)
	for totalLength > 0 {
		blockLength := int(binary.LittleEndian.Uint16(f.fixed(2, false)))
		if f.err != nil {
			return nil
		}
		if blockLength > totalLength {
			f.err = fmt.Errorf("block length %d exceeds total length %d", blockLength, totalLength)
			return nil
		}
		buf = append(buf, f.fixed(blockLength, false))
		if f.err != nil {
			return nil
		}
		totalLength -= blockLength
	}

	return buf
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
