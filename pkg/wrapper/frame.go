package wrapper

import (
	"encoding/binary"
	"fmt"
)

const (
	// Version is the WRAPPER protocol version.
	Version uint16 = 1
)

// Frame represents a WRAPPER frame.
type Frame struct {
	Version    uint16
	SrcAddr    uint16
	DstAddr    uint16
	Length     uint16
	Payload    []byte
}

// Encode serializes the frame into a byte slice.
func (f *Frame) Encode() ([]byte, error) {
	if len(f.Payload) != int(f.Length) {
		return nil, fmt.Errorf("payload length does not match length field")
	}
	buf := make([]byte, 8+f.Length)
	binary.BigEndian.PutUint16(buf[0:2], f.Version)
	binary.BigEndian.PutUint16(buf[2:4], f.SrcAddr)
	binary.BigEndian.PutUint16(buf[4:6], f.DstAddr)
	binary.BigEndian.PutUint16(buf[6:8], f.Length)
	copy(buf[8:], f.Payload)
	return buf, nil
}

// Decode deserializes a byte slice into a frame.
func (f *Frame) Decode(src []byte) error {
	if len(src) < 8 {
		return fmt.Errorf("insufficient data for frame header")
	}
	f.Version = binary.BigEndian.Uint16(src[0:2])
	f.SrcAddr = binary.BigEndian.Uint16(src[2:4])
	f.DstAddr = binary.BigEndian.Uint16(src[4:6])
	f.Length = binary.BigEndian.Uint16(src[6:8])
	if len(src) < 8+int(f.Length) {
		return fmt.Errorf("insufficient data for frame payload")
	}
	f.Payload = src[8 : 8+f.Length]
	return nil
}
