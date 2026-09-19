package rpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var defaultTransportParameters = TransportParameters{packetSize: 1024, maxBytes: 1024, windowSize: 16, ackBehind: 1, ackTimeout: 600}

const idleTimeout = 30 * time.Minute

type connectionSet map[net.Conn]struct{}

type Server struct {
	conns   connectionSet
	mu      sync.RWMutex
	wg      sync.WaitGroup
	closing bool
	nextID  int
}

type Session struct {
	userID string
}

// encoding functions (abstract later)

func NewServer() *Server {
	return &Server{
		conns: connectionSet{},
		mu:    sync.RWMutex{},
		wg:    sync.WaitGroup{},
	}
}

func (s *Server) Serve(ctx context.Context, l net.Listener) error {
	go func() {
		<-ctx.Done()
		l.Close()
	}()
	for {
		fmt.Println("standing by for connections")
		conn, err := l.Accept()
		if err != nil {
			// context is closed, time to close up shop
			if ctx.Err() != nil {
				return nil
			}
			/* otherwise, it's probably a recoverable error
			wait a bit to avoid a hot loop and then retry
			*/
			time.Sleep(100 * time.Millisecond)
		}
		id, ok := s.add(conn)
		if !ok {
			conn.Close()
			continue
		}
		go s.handle(conn, id)
	}
}

func (s *Server) handle(c net.Conn, id int) error {
	fmt.Println("new connection")
	defer s.remove(c)
	session := Session{}
	for {
		c.SetReadDeadline(time.Now().Add(idleTimeout))
		packet, err := readStraightPacket(c)
		fmt.Println("got a read")
		if err != nil {
			fmt.Printf("packet read error %v\n", err)
			return err
		}
		if p, err := unmarshalPipeFrame(packet); err != nil {
			fmt.Printf("unmarshaling error %v\n", err)
			return err
		} else {
			res, err := dispatchPipeMessage(&session, p)
			if err != nil {
				fmt.Printf("dispatch: %v\n", err)
			} else {
				fmt.Println("sending response")
				err := writeStraightPacket(c, res)
				if err != nil {
					fmt.Printf("write packet error %v", err)
				} else {
					fmt.Println("sent successful")
				}
			}
		}
	}
}

func (s *Server) remove(c net.Conn) {
	fmt.Println("removing connection")
	s.mu.Lock()
	delete(s.conns, c)
	s.mu.Unlock()
	c.Close()
	s.wg.Done()
}

func (s *Server) add(c net.Conn) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return 0, false
	}
	s.conns[c] = struct{}{}
	s.wg.Add(1)
	return s.nextID, true
}

func readStraightPacket(r io.Reader) ([]byte, error) {
	buf := make([]byte, 3)
	fmt.Println("boutta read")
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	fmt.Println("read them bytes")

	length := binary.LittleEndian.Uint16(buf)
	if length < 5 {
		return nil, fmt.Errorf("packet too short: %d", length)
	}

	fmt.Printf("will expect %d bytes\n", length)

	buf = make([]byte, int(length)-len(buf))

	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	fmt.Println("bytes read!")

	return buf, nil
}

func writeStraightPacket(w io.Writer, a Appendable) error {
	msgSize := a.Size()
	totalSize := msgSize + 3
	fmt.Printf("got total size of %d\n", totalSize)
	buf := make([]byte, 0, totalSize)
	buf = binary.LittleEndian.AppendUint16(buf, uint16(totalSize))
	buf = append(buf, 0)
	buf = a.Append(buf)
	n, err := w.Write(buf)
	fmt.Printf("wrote %d bytes (%x)\n", n, buf)
	return err
}
