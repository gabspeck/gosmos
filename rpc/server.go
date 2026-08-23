package rpc

import (
	"gabriels.io/gosmos/svc"
)

const RoutingControlFrame = 0xffff

const ReceiverControlFrame = "ControlFrame"

const TypeControlFrameConnectionEstablished = 0x4

const FunctionControlFrameConnectionEstablished = "ConnectionEstablished"

var x svc.ControlFrameX

type (
	RoutingToReceiver      = map[uint16]string
	ReceiverTypeToFunction = map[string]map[uint8]string
)

var receivers = RoutingToReceiver{
	RoutingControlFrame: ReceiverControlFrame,
}

var functions = ReceiverTypeToFunction{
	ReceiverControlFrame: {
		TypeControlFrameConnectionEstablished: FunctionControlFrameConnectionEstablished,
	},
}
