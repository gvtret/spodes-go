package transport

import "net"

// Transport defines a common interface for different transport layers,
// such as HDLC or TCP/IP WRAPPER.
type Transport interface {
	// Connect initiates the connection to the remote endpoint.
	// It returns the initial message to be sent, if any, or an error.
	Connect() ([]byte, error)

	// Disconnect terminates the connection.
	// It returns the final message to be sent, if any, or an error.
	Disconnect() ([]byte, error)

	// IsConnected returns true if the transport layer is in a connected state.
	IsConnected() bool

	// Send transmits a PDU (Protocol Data Unit) over the transport.
	// It returns a slice of byte slices, representing one or more frames to be sent.
	Send(pdu []byte) ([][]byte, error)

	// Receive processes an incoming byte stream, which may contain partial or multiple frames.
	// It returns a slice of byte slices, representing any response frames to be sent.
	Receive(src []byte) ([][]byte, error)

	// Read blocks until a complete PDU has been received and reassembled.
	// It returns the reassembled PDU, the source address of the sender, or an error.
	Read() ([]byte, net.Addr, error)
}
