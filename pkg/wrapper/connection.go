package wrapper

import (
	"fmt"
	"io"
	"net"
)

// Conn wraps a net.Conn to send and receive WRAPPER frames.
type Conn struct {
	conn net.Conn
}

// NewConn creates a new Conn.
func NewConn(conn net.Conn) *Conn {
	return &Conn{conn: conn}
}

// Send sends a WRAPPER frame.
func (c *Conn) Send(frame *Frame) error {
	encoded, err := frame.Encode()
	if err != nil {
		return err
	}
	_, err = c.conn.Write(encoded)
	return err
}

// Receive receives a WRAPPER frame.
func (c *Conn) Receive() (*Frame, error) {
	header := make([]byte, 8)
	_, err := io.ReadFull(c.conn, header)
	if err != nil {
		return nil, err
	}

	frame := &Frame{}
	frame.Version = uint16(header[0])<<8 | uint16(header[1])
	frame.SrcAddr = uint16(header[2])<<8 | uint16(header[3])
	frame.DstAddr = uint16(header[4])<<8 | uint16(header[5])
	frame.Length = uint16(header[6])<<8 | uint16(header[7])

	if frame.Version != Version {
		return nil, fmt.Errorf("invalid wrapper version: %d", frame.Version)
	}

	payload := make([]byte, frame.Length)
	_, err = io.ReadFull(c.conn, payload)
	if err != nil {
		return nil, err
	}
	frame.Payload = payload

	return frame, nil
}
