package rpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/rpc"
	"reflect"
	"sync"
	"time"
)

var planCache sync.Map

type planField struct {
	index  int
	name   string
	decode fieldDecoder
}

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
		f.err = fmt.Errorf("field %d: delimiter 0x%02x not found", term)
		return nil
	}
	v := f.buf[:termIdx] // skip terminator
	f.buf, f.fieldIndex = f.buf[termIdx:], f.fieldIndex+1
	return v
}

func (f *fieldReader) Uint16() uint16 {
	b := f.tagged(TagRequestUint16, 1)
	if f.err != nil {
		return 0
	}
	return binary.BigEndian.Uint16(b)
}

func (f *fieldReader) Uint32() uint32 {
	b := f.tagged(TagRequestUint32, 1)
	if f.err != nil {
		return 0
	}
	return binary.BigEndian.Uint32(b)
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

func planFor(t reflect.Type) ([]planField, error) {
	if p, ok := planCache.Load(t); ok {
		return p.([]planField), nil
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is not a struct", t)
	}
	var p []planField
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Tag.Get("rpc") == "-" {
			continue
		}
		dec, err := decoderFor(f.Type)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", t.Name(), f.Name, err)
		}
		p = append(p, planField{index: i, name: f.Name, decode: dec})
	}
	planCache.Store(t, p)
	return p, nil
}

type fieldDecoder func(reflect.Value, *fieldReader)

func decoderFor(t reflect.Type) (fieldDecoder, error) {
	switch t.Kind() {
	case reflect.Uint8:
		return func(v reflect.Value, f *fieldReader) { v.SetUint(uint64(f.Byte())) }, nil
	case reflect.Uint16:
		return func(v reflect.Value, f *fieldReader) { v.SetUint(uint64(f.Uint16())) }, nil
	case reflect.Uint32:
		return func(v reflect.Value, f *fieldReader) { v.SetUint(uint64(f.Uint32())) }, nil
	case reflect.String:
		return func(v reflect.Value, f *fieldReader) { v.SetString(f.NarrowString()) }, nil
	}

	return nil, fmt.Errorf("unsupported field type %s", t)
}

func decodeFields(x any, f *fieldReader) error {
	rv := reflect.ValueOf(x)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("fields must be a non-nil pointer, got %T", x)
	}
	rv = rv.Elem()
	plan, err := planFor(rv.Type())
	if err != nil {
		return err
	}
	for _, pf := range plan {
		pf.decode(rv.Field(pf.index), f)
		if f.err != nil {
			return fmt.Errorf("%s.%s: %w", rv.Type().Name(), pf.name, f.err)
		}
	}
	return f.Done()
}

type MosServerCodec struct {
	conn net.Conn
	seq  uint64
	buf  []byte
}

const timeout = time.Second * 5

func NewMosServerCodec(conn net.Conn) *MosServerCodec {
	return &MosServerCodec{
		conn: conn,
		buf:  make([]byte, 2<<15),
	}
}

func (c *MosServerCodec) ReadRequestHeader(r *rpc.Request) error {
	r.Seq = c.seq
	c.seq += 1

	if err := c.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
}

func (c *MosServerCodec) ReadRequestBody(b any) error {
	if b == nil {
		return nil
	}

	return nil
}

func (c *MosServerCodec) WriteResponse(r *rpc.Response, x any) error {
	fmt.Printf("wr: %v; %v\n", r, x)
	return nil
}

func (c *MosServerCodec) Close() error {
	return nil
}
