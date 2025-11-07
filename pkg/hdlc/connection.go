package hdlc

import (
	"bytes"
	"sync"
	"time"
)

// HDLCError represents an HDLC-specific error
type HDLCError struct {
	Message    string
	ShouldExit bool // Indicates if the connection should be terminated
}

func (e *HDLCError) Error() string {
	return e.Message
}

// Predefined HDLC errors
var (
	ErrNotConnected              = &HDLCError{Message: "not connected", ShouldExit: true}
	ErrAlreadyConnected          = &HDLCError{Message: "already connected or connecting"}
	ErrInvalidUA                 = &HDLCError{Message: "did not receive UA in response to SNRM"}
	ErrAckTimeout                = &HDLCError{Message: "ack timeout"}
	ErrInactivityTimeout         = &HDLCError{Message: "inactivity timeout", ShouldExit: true}
	ErrUnexpectedFrame           = &HDLCError{Message: "unexpected frame"}
	ErrInvalidFrame              = &HDLCError{Message: "invalid frame"}
	ErrConnectionTerminated      = &HDLCError{Message: "connection terminated", ShouldExit: true}
	ErrUnexpectedDisconnect      = &HDLCError{Message: "unexpected disconnect", ShouldExit: true}
	ErrFrameRejected             = &HDLCError{Message: "frame rejected", ShouldExit: true}
	ErrDestinationAddressMissing = &HDLCError{Message: "destination address is missing"}
	ErrSourceAddressMissing      = &HDLCError{Message: "source address is missing"}
)

// Define connection states
const (
	StateDisconnected = "disconnected"
	StateConnecting   = "connecting"
	StateConnected    = "connected"
)

// HDLCConnection manages the HDLC connection
type HDLCConnection struct {
	state             string
	destAddr          []byte
	srcAddr           []byte
	sendSeq           uint8
	recvSeq           uint8
	lastAckedSeq      uint8
	windowSize        int
	sentFrames        map[uint8]*HDLCFrame
	recvBuffer        map[uint8]*HDLCFrame
	mutex             sync.Mutex
	ackChannel        chan uint8
	inactivityTimeout time.Duration
	lastActivity      time.Time
	readBuffer        bytes.Buffer
	inFrame           bool
}

// SetAddress sets the destination and source addresses for the connection
func (c *HDLCConnection) SetAddress(dest, src []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.destAddr = dest
	c.srcAddr = src
}

// NewHDLCConnection creates a new HDLC connection
func NewHDLCConnection() *HDLCConnection {
	return &HDLCConnection{
		state:             StateDisconnected,
		windowSize:        MaxWindowSize,
		sentFrames:        make(map[uint8]*HDLCFrame),
		recvBuffer:        make(map[uint8]*HDLCFrame),
		ackChannel:        make(chan uint8, 1),
		inactivityTimeout: time.Duration(InactivityTimeout) * time.Millisecond,
		readBuffer:        bytes.Buffer{},
		inFrame:           false,
	}
}

// Connect generates an SNRM frame to initiate a connection
func (c *HDLCConnection) Connect() ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.state != StateDisconnected {
		return nil, ErrAlreadyConnected
	}
	if len(c.destAddr) == 0 {
		return nil, ErrDestinationAddressMissing
	}
	if len(c.srcAddr) == 0 {
		return nil, ErrSourceAddressMissing
	}

	c.state = StateConnecting
	snrmFrame := &HDLCFrame{DA: c.destAddr, SA: c.srcAddr, Control: UFrameSNRM, PF: true}
	return EncodeFrame(snrmFrame.DA, snrmFrame.SA, snrmFrame.Control, snrmFrame.Information, snrmFrame.Segmented)
}

// HandleFrame processes an incoming HDLC frame and returns the response frame
func (c *HDLCConnection) HandleFrame(frameData []byte) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	frame, err := DecodeFrame(frameData, 0)
	if err != nil {
		return nil, err
	}

	switch c.state {
	case StateDisconnected:
		if frame.Control == UFrameSNRM {
			c.state = StateConnected // Server moves to connected state after sending UA
			c.lastActivity = time.Now()
			uaFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeU, Control: UFrameUA, PF: true}
			return EncodeFrame(uaFrame.DA, uaFrame.SA, uaFrame.Control, uaFrame.Information, uaFrame.Segmented)
		}
	case StateConnecting:
		if frame.Control == UFrameUA {
			c.state = StateConnected
			c.lastActivity = time.Now()
			return nil, nil // No response needed
		}
		return nil, ErrInvalidUA
	case StateConnected:
		if frame.Control == UFrameUA {
			// This is in response to a DISC
			c.state = StateDisconnected
			return nil, nil
		}
		// Handle data and supervisory frames
		return c.handleConnectedState(frame)
	}

	return nil, ErrNotConnected
}

// handleConnectedState processes frames when in a connected state
func (c *HDLCConnection) handleConnectedState(frame *HDLCFrame) ([]byte, error) {
	switch frame.Type {
	case FrameTypeI:
		if frame.NS != c.recvSeq {
			rejFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeS, Control: SFrameREJ | (c.recvSeq << 5)}
			return EncodeFrame(rejFrame.DA, rejFrame.SA, rejFrame.Control, rejFrame.Information, rejFrame.Segmented)
		}
		c.recvSeq = (c.recvSeq + 1) % 8
		rrFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeS, Control: SFrameRR | (c.recvSeq << 5)}
		return EncodeFrame(rrFrame.DA, rrFrame.SA, rrFrame.Control, rrFrame.Information, rrFrame.Segmented)
	case FrameTypeU:
		if frame.Control == UFrameDISC {
			c.state = StateDisconnected
			uaFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeU, Control: UFrameUA, PF: true}
			return EncodeFrame(uaFrame.DA, uaFrame.SA, uaFrame.Control, uaFrame.Information, uaFrame.Segmented)
		}
	}
	return nil, nil // No response for other frames for now
}

// SendData generates an I-frame for the given data payload
func (c *HDLCConnection) SendData(data []byte) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.state != StateConnected {
		return nil, ErrNotConnected
	}

	frame := &HDLCFrame{
		DA:          c.destAddr,
		SA:          c.srcAddr,
		Type:        FrameTypeI,
		NS:          c.sendSeq,
		NR:          c.recvSeq,
		Information: data,
	}
	frame.Control = (frame.NS << 1) | (frame.NR << 5)

	c.sendSeq = (c.sendSeq + 1) % 8
	return EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, frame.Segmented)
}

// Disconnect generates a DISC frame to terminate the connection
func (c *HDLCConnection) Disconnect() ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.state != StateConnected {
		return nil, ErrNotConnected
	}
	if len(c.destAddr) == 0 {
		return nil, ErrDestinationAddressMissing
	}
	if len(c.srcAddr) == 0 {
		return nil, ErrSourceAddressMissing
	}

	discFrame := &HDLCFrame{DA: c.destAddr, SA: c.srcAddr, Control: UFrameDISC, PF: true}
	return EncodeFrame(discFrame.DA, discFrame.SA, discFrame.Control, discFrame.Information, discFrame.Segmented)
}

// Handle processes an incoming byte stream and returns any response frames
func (c *HDLCConnection) Handle(data []byte) ([][]byte, error) {
	c.readBuffer.Write(data)
	var responses [][]byte

	for {
		idx := bytes.IndexByte(c.readBuffer.Bytes(), FlagByte)
		if idx == -1 {
			break // No flag found
		}

		if idx > 0 {
			// There is data before the flag, which could be the end of a previous frame
			frameData := make([]byte, idx)
			c.readBuffer.Read(frameData)
			fullFrame := append([]byte{FlagByte}, frameData...)
			fullFrame = append(fullFrame, FlagByte)

			response, err := c.HandleFrame(fullFrame)
			if err == nil && response != nil {
				responses = append(responses, response)
			}
		}
		c.readBuffer.Next(1) // Consume the flag
	}
	return responses, nil
}

// IsConnected returns true if the connection is in the Connected state
func (c *HDLCConnection) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.state == StateConnected
}

// SetState sets the connection state
func (c *HDLCConnection) SetState(state string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.state = state
}
