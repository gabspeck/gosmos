package rpc

import "testing"

func TestControlFrameUnmarshal(t *testing.T) {
	buf := []byte{0x04}

	cf := ControlFrame{}
	err := cf.UnmarshalBinary(buf)
	if err != nil {
		t.Fatal(err)
	}

	if cf.Type != 0x04 {
		t.Fatalf("unexpected type: %d", cf.Type)
	}
}
