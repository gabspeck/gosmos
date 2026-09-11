package rpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

var defaultTransportParameters = TransportParameters{packetSize: 1024, maxBytes: 1024, windowSize: 16, ackBehind: 1, ackTimeout: 600}

type Server struct {
	ctx  context.Context
	conn io.ReadWriteCloser
}

func NewServer(conn io.ReadWriteCloser) *Server {
	return &Server{
		conn: conn,
	}
}

func (s *Server) HandleConnection(c io.ReadWriteCloser) error {
	buf, err := readStraightPacket(c)
	if err != nil {
		return err
	}
	m, err := unmarshalPipeMessage(buf)
	if err != nil {
		return nil
	}

	return s.handlePipeMessage(m)
}

func readStraightPacket(r io.Reader) ([]byte, error) {
	buf := make([]byte, 3)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	length := binary.LittleEndian.Uint16(buf)
	if length < 5 {
		return nil, fmt.Errorf("packet too short: %d", length)
	}

	buf = make([]byte, length)

	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func writeStraightPacket(w io.Writer) error {
	return nil
}

func (s *Server) handlePipeMessage(m PipeMessage) error {
	switch m := m.(type) {
	case ControlFrame:
		return s.handleControlFrame(m)
	}
	return fmt.Errorf("unhandled pipe message %q", reflect.TypeOf(m))
}

func (s *Server) handleControlFrame(c ControlFrame) error {
	switch c.(type) {
	case *ConnectionEstablished:
		return s.sendTransportParameters()
	}
	return fmt.Errorf("unhandled control frame %q", reflect.TypeOf(c))
}

func (s *Server) sendTransportParameters() error {
}
