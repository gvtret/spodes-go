package hdlc

import (
	"net"
	"time"

	"github.com/tarm/serial"
)

// Transport defines the interface for HDLC transport layer
type Transport interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
	SetReadDeadline(t time.Time) error
}

// TCPTransport implements Transport for TCP connections
type TCPTransport struct {
	conn net.Conn
}

// NewTCPTransport creates a new TCPTransport
func NewTCPTransport(conn net.Conn) *TCPTransport {
	return &TCPTransport{conn: conn}
}

// Read reads data from the TCP connection
func (t *TCPTransport) Read(b []byte) (int, error) {
	return t.conn.Read(b)
}

// Write writes data to the TCP connection
func (t *TCPTransport) Write(b []byte) (int, error) {
	return t.conn.Write(b)
}

// Close closes the TCP connection
func (t *TCPTransport) Close() error {
	return t.conn.Close()
}

// SetReadDeadline sets the read deadline for the TCP connection
func (t *TCPTransport) SetReadDeadline(tm time.Time) error {
	return t.conn.SetReadDeadline(tm)
}

// UDPTransport implements Transport for UDP connections
type UDPTransport struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

// NewUDPTransport creates a new UDPTransport
func NewUDPTransport(conn *net.UDPConn, addr *net.UDPAddr) *UDPTransport {
	return &UDPTransport{conn: conn, addr: addr}
}

// Read reads data from the UDP connection
func (t *UDPTransport) Read(b []byte) (int, error) {
	return t.conn.Read(b)
}

// Write writes data to the UDP connection
func (t *UDPTransport) Write(b []byte) (int, error) {
	return t.conn.WriteToUDP(b, t.addr)
}

// Close closes the UDP connection
func (t *UDPTransport) Close() error {
	return t.conn.Close()
}

// SetReadDeadline sets the read deadline for the UDP connection
func (t *UDPTransport) SetReadDeadline(tm time.Time) error {
	return t.conn.SetReadDeadline(tm)
}

// SerialTransport implements Transport for serial port connections
type SerialTransport struct {
	port *serial.Port
}

// NewSerialTransport creates a new SerialTransport
func NewSerialTransport(portName string) (*SerialTransport, error) {
	config := &serial.Config{
		Name:   portName,
		Baud:   9600, // Default baud rate as per СТО 34.01-5.1-006-2023
		Parity: serial.ParityNone,
	}
	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, err
	}
	return &SerialTransport{port: port}, nil
}

// Read reads data from the serial port
func (t *SerialTransport) Read(b []byte) (int, error) {
	return t.port.Read(b)
}

// Write writes data to the serial port
func (t *SerialTransport) Write(b []byte) (int, error) {
	return t.port.Write(b)
}

// Close closes the serial port
func (t *SerialTransport) Close() error {
	return t.port.Close()
}

// SetReadDeadline sets the read deadline for the serial port (not supported)
func (t *SerialTransport) SetReadDeadline(tm time.Time) error {
	// Serial ports typically do not support deadlines; use timeouts in application logic
	return nil
}
