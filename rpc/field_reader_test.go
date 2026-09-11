package rpc

import (
	"fmt"
	"testing"
)

func TestVLI(t *testing.T) {
	testCases := []struct {
		desc string
		val  []byte
		want uint32
		err  error
	}{
		{
			desc: "1-byte minimum",
			val:  []byte{0b00000000},
			want: 0,
		},
		{
			desc: "1-byte maximum",
			val:  []byte{0b00111111},
			want: 63,
		},
		{
			desc: "2-byte minimum",
			val:  []byte{0b10000000, 0b01000000},
			want: 64,
		},
		{
			desc: "2-byte maximum",
			val:  []byte{0b10111111, 0b11111111},
			want: 16_383,
		},
		{
			desc: "4-byte minimum",
			val:  []byte{0b11000000, 0b00000000, 0b01000000, 0b00000000},
			want: 16_384,
		},
		{
			desc: "4-byte maximum",
			val:  []byte{0b11111111, 0b11111111, 0b11111111, 0b11111111},
			want: 1_073_741_823,
		},
		{
			desc: "invalid tag",
			val:  []byte{0b01111111},
			want: 0,
			err:  fmt.Errorf("invalid VLI form bits: 0b01"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			f := fieldReader{buf: tC.val}
			actual := f.VLI()
			if actual != tC.want {
				t.Fatalf("wanted %d, got %d", tC.want, actual)
			}
			if fmt.Sprintf("%v", tC.err) != fmt.Sprintf("%v", f.err) {
				t.Fatalf("wanted error %q; got %q", tC.err, f.err)
			}
		})
	}
}

func TestVSize(t *testing.T) {
	testCases := []struct {
		desc string
		val  []byte
		want uint32
		err  error
	}{
		{
			desc: "1-byte mininum",
			val:  []byte{0b1000_0000},
			want: 0,
			err:  nil,
		},
		{
			desc: "1-byte maximum",
			val:  []byte{0b1111_1111},
			want: 127,
			err:  nil,
		},
		{
			desc: "2-byte mininum",
			val:  []byte{0b0000_0000, 0b1000_0000},
			want: 128,
			err:  nil,
		},
		{
			desc: "2-byte maximum",
			val:  []byte{0b0111_1111, 0b1111_1111},
			want: 32767,
			err:  nil,
		},
		{
			desc: "2-byte with 1-byte encodable",
			val:  []byte{0b0000_0000, 0b0000_0000},
			want: 0,
			err:  nil,
		},
		{
			desc: "2-byte tag, buffer overrun",
			val:  []byte{0b0111_1111},
			want: 0,
			err:  fmt.Errorf("field 0: truncated"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			f := fieldReader{buf: tC.val}
			actual := f.vsize()
			if actual != tC.want {
				t.Fatalf("wanted %d, got %d", tC.want, actual)
			}
			if fmt.Sprintf("%v", tC.err) != fmt.Sprintf("%v", f.err) {
				t.Fatalf("wanted error %q; got %q", tC.err, f.err)
			}
		})
	}
}
