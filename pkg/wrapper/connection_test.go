package wrapper

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnSendReceive(t *testing.T) {
	// Create a mock TCP connection
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Create a server-side Conn
	serverConn := NewConn(server)

	// Create a client-side Conn
	clientConn := NewConn(client)

	// Create a frame to send
	frameToSend := &Frame{
		Version: Version,
		SrcAddr: 1,
		DstAddr: 2,
		Length:  5,
		Payload: []byte("hello"),
	}

	// Send the frame from the client
	go func() {
		clientConn.Send(frameToSend)
	}()

	// Receive the frame on the server
	receivedFrame, err := serverConn.Receive()
	assert.NoError(t, err)

	// Check that the received frame is correct
	assert.Equal(t, frameToSend.Version, receivedFrame.Version)
	assert.Equal(t, frameToSend.SrcAddr, receivedFrame.SrcAddr)
	assert.Equal(t, frameToSend.DstAddr, receivedFrame.DstAddr)
	assert.Equal(t, frameToSend.Length, receivedFrame.Length)
	assert.Equal(t, frameToSend.Payload, receivedFrame.Payload)
}

func TestConnReceiveInvalidVersion(t *testing.T) {
	// Create a mock TCP connection
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Create a server-side Conn
	serverConn := NewConn(server)

	// Send an invalid frame from the client
	go func() {
		invalidFrame := &Frame{
			Version: 999, // Invalid version
			SrcAddr: 1,
			DstAddr: 2,
			Length:  5,
			Payload: []byte("hello"),
		}
		encoded, _ := invalidFrame.Encode()
		client.Write(encoded)
	}()

	_, err := serverConn.Receive()
	assert.Error(t, err)
}
