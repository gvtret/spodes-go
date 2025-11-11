package wrapper

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockConn is a mock net.Conn for testing.
type mockConn struct {
	net.Conn
	readBuffer  bytes.Buffer
	writeBuffer bytes.Buffer
	mutex       sync.Mutex
	closed      bool
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return 0, net.ErrClosed
	}
	return c.readBuffer.Read(b)
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return 0, net.ErrClosed
	}
	return c.writeBuffer.Write(b)
}

func (c *mockConn) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.closed = true
	return nil
}

func (c *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 12345,
	}
}

func TestConnectionSendAndReceive(t *testing.T) {
	mock := &mockConn{}
	config := DefaultConfig()
	conn := NewConnection(mock, config)

	pduToSend := []byte("hello world")

	// Test Send
	frames, err := conn.Send(pduToSend)
	assert.NoError(t, err)
	assert.Len(t, frames, 1)

	// Simulate sending the frame over the network
	mock.readBuffer.Write(frames[0])

	// Test Receive
	_, err = conn.Receive(mock.readBuffer.Bytes())
	assert.NoError(t, err)

	// Test Read
	receivedPDU, addr, err := conn.Read()
	assert.NoError(t, err)
	assert.Equal(t, pduToSend, receivedPDU)
	assert.NotNil(t, addr)
	assert.Equal(t, "127.0.0.1:12345", addr.String())
}

func TestConnectionReadTimeout(t *testing.T) {
	mock := &mockConn{}
	config := DefaultConfig()
	config.ReadTimeout = 50 * time.Millisecond
	conn := NewConnection(mock, config)

	_, _, err := conn.Read()
	assert.Error(t, err)
	assert.Equal(t, "read timeout", err.Error())
}

func TestConnectionInvalidFrame(t *testing.T) {
	mock := &mockConn{}
	config := DefaultConfig()
	conn := NewConnection(mock, config)

	// Send a corrupted frame (e.g., wrong version)
	invalidFrame := &Frame{
		Version: 999, // Invalid
		SrcAddr: config.SrcAddr,
		DstAddr: config.DstAddr,
		Length:  4,
		Payload: []byte("test"),
	}
	encoded, _ := invalidFrame.Encode()
	mock.readBuffer.Write(encoded)

	// The Receive method should handle this gracefully
	_, err := conn.Receive(mock.readBuffer.Bytes())
	assert.NoError(t, err)

	// Since the frame was invalid, no PDU should be available to read.
	// We expect a timeout here.
	config.ReadTimeout = 50 * time.Millisecond
	_, _, err = conn.Read()
	assert.Error(t, err)
}
