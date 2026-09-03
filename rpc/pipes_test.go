package rpc

// import (
// 	"encoding/hex"
// 	"reflect"
// 	"testing"
// )

// func TestControlFrameUnmarshal(t *testing.T) {
// 	buf := []byte{0x04}

// 	cf := controlFrameMessage{}
// 	err := cf.UnmarshalBinary(buf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if cf.Type != 0x04 {
// 		t.Fatalf("unexpected type: %d", cf.Type)
// 	}
// }

// func TestPipeOpenUnmarshal(t *testing.T) {
// 	buf, _ := hex.DecodeString("000001004c4f4753525600550006000000")

// 	expected := PipeOpenRequest{
// 		reserved:    0,
// 		pipeIndex:   1,
// 		serviceName: "LOGSRV",
// 		parameter:   "U",
// 		version:     6,
// 	}

// 	po := PipeOpenRequest{}
// 	err := po.UnmarshalBinary(buf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !reflect.DeepEqual(expected, po) {
// 		t.Fatalf("sructs differ %v\n%v", buf, po)
// 	}
// }
