package mos

import (
	"context"
	"io"
)

type Server struct {
	ctx context.Context
}

func NewServer() *Server {
	return &Server{}
}

type PipeMessage interface {
	cmd() byte
	routing() uint16
}

type ControlBody interface {
	PipeMessage
	controlType() byte
}

func (s *Server) HandleConnection(c io.ReadWriteCloser) {
	// get package length and read into buffer
	// read routing value
}
