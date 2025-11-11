package hdlc

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"
)

// Config holds the configuration parameters for an HDLC connection.
type Config struct {
	WindowSize           int
	MaxFrameSize         int
	InactivityTimeout    time.Duration
	FrameAssemblyTimeout time.Duration
	RetransmissionTimeout time.Duration
	DestAddr             []byte
	SrcAddr              []byte
}

// DefaultConfig returns a new Config object with default values.
func DefaultConfig() *Config {
	return &Config{
		WindowSize:           MaxWindowSize,
		MaxFrameSize:         128,
		InactivityTimeout:    time.Duration(InactivityTimeout) * time.Millisecond,
		FrameAssemblyTimeout: 2 * time.Second,
		RetransmissionTimeout: 5 * time.Second,
	}
}

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
	state                string
	destAddr             []byte
	srcAddr              []byte
	sendSeq              uint8
	recvSeq              uint8
	lastAckedSeq         uint8
	windowSize           int
	maxFrameSize         int
	sentFrames           map[uint8]*HDLCFrame
	sentTimes            map[uint8]time.Time
	recvBuffer           map[uint8]*HDLCFrame
	segmentBuffer        []byte
	ReassembledData      chan []byte
	RetransmitFrames     chan []byte
	mutex                sync.Mutex
	ackChannel           chan uint8
	isPeerReceiverReady  bool
	inactivityTimeout    time.Duration
	frameAssemblyTimeout time.Duration
	retransmissionTimeout time.Duration
	lastActivity         time.Time
	readBuffer           bytes.Buffer
}

// NewHDLCConnection creates a new HDLC connection with the given configuration.
// If config is nil, default configuration is used.
func NewHDLCConnection(config *Config) *HDLCConnection {
	if config == nil {
		config = DefaultConfig()
	}
	conn := &HDLCConnection{
		state:                StateDisconnected,
		windowSize:           config.WindowSize,
		maxFrameSize:         config.MaxFrameSize,
		inactivityTimeout:    config.InactivityTimeout,
		frameAssemblyTimeout: config.FrameAssemblyTimeout,
		retransmissionTimeout: config.RetransmissionTimeout,
		destAddr:             config.DestAddr,
		srcAddr:              config.SrcAddr,
		sentFrames:           make(map[uint8]*HDLCFrame),
		sentTimes:            make(map[uint8]time.Time),
		recvBuffer:           make(map[uint8]*HDLCFrame),
		segmentBuffer:        make([]byte, 0),
		ReassembledData:      make(chan []byte, 10),
		RetransmitFrames:     make(chan []byte, 10),
		ackChannel:           make(chan uint8, 1),
		isPeerReceiverReady:  true,
		readBuffer:           bytes.Buffer{},
	}
	go conn.retransmissionDaemon()
	return conn
}

func (c *HDLCConnection) retransmissionDaemon() {
	ticker := time.NewTicker(c.retransmissionTimeout / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		for ns, t := range c.sentTimes {
			if time.Since(t) > c.retransmissionTimeout {
				if frameToResend, ok := c.sentFrames[ns]; ok {
					encodedFrame, err := EncodeFrame(frameToResend.DA, frameToResend.SA, frameToResend.Control, frameToResend.Information, frameToResend.Segmented)
					if err == nil {
						// Non-blocking send to avoid deadlock
						select {
						case c.RetransmitFrames <- encodedFrame:
						default:
						}
					}
				}
			}
		}
		c.mutex.Unlock()
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

// HandleFrame processes a decoded HDLC frame and returns the response frame
func (c *HDLCConnection) HandleFrame(frame *HDLCFrame) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lastActivity = time.Now()

	switch c.state {
	case StateDisconnected:
		if frame.Control == UFrameSNRM {
			c.state = StateConnected
			uaFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeU, Control: UFrameUA, PF: true}
			return EncodeFrame(uaFrame.DA, uaFrame.SA, uaFrame.Control, uaFrame.Information, uaFrame.Segmented)
		}
	case StateConnecting:
		if frame.Control == UFrameUA {
			c.state = StateConnected
			return nil, nil
		}
		return nil, ErrInvalidUA
	case StateConnected:
		if frame.Control == UFrameUA {
			c.state = StateDisconnected
			return nil, nil
		}
		return c.handleConnectedState(frame)
	}

	return nil, ErrNotConnected
}

// handleConnectedState processes frames when in a connected state
func (c *HDLCConnection) handleConnectedState(frame *HDLCFrame) ([]byte, error) {
	switch frame.Type {
	case FrameTypeI:
		// Buffer out-of-order frames
		if frame.NS != c.recvSeq {
			if _, exists := c.recvBuffer[frame.NS]; !exists {
				c.recvBuffer[frame.NS] = frame
			}
			// Send SREJ for the missing frame
			srejFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeS, Control: SFrameSREJ | (c.recvSeq << 5)}
			return EncodeFrame(srejFrame.DA, srejFrame.SA, srejFrame.Control, srejFrame.Information, srejFrame.Segmented)
		}

		// Process the current in-order frame
		c.segmentBuffer = append(c.segmentBuffer, frame.Information...)
		if !frame.Segmented {
			if c.ReassembledData != nil {
				c.ReassembledData <- c.segmentBuffer
			}
			c.segmentBuffer = make([]byte, 0)
		}
		c.recvSeq = (c.recvSeq + 1) % 8

		// Process any buffered frames that are now in order
		for {
			if bufferedFrame, ok := c.recvBuffer[c.recvSeq]; ok {
				c.segmentBuffer = append(c.segmentBuffer, bufferedFrame.Information...)
				if !bufferedFrame.Segmented {
					if c.ReassembledData != nil {
						c.ReassembledData <- c.segmentBuffer
					}
					c.segmentBuffer = make([]byte, 0)
				}
				delete(c.recvBuffer, c.recvSeq)
				c.recvSeq = (c.recvSeq + 1) % 8
			} else {
				break
			}
		}

		rrFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeS, Control: SFrameRR | (c.recvSeq << 5)}
		return EncodeFrame(rrFrame.DA, rrFrame.SA, rrFrame.Control, rrFrame.Information, rrFrame.Segmented)

	case FrameTypeU:
		if frame.Control == UFrameDISC {
			c.state = StateDisconnected
			uaFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeU, Control: UFrameUA, PF: true}
			return EncodeFrame(uaFrame.DA, uaFrame.SA, uaFrame.Control, uaFrame.Information, uaFrame.Segmented)
		}
	case FrameTypeS:
		nr := (frame.Control >> 5) & 0x07

		switch frame.Control & 0x0F {
		case SFrameRR, SFrameREJ, SFrameSREJ:
			c.lastAckedSeq = nr
			for i := c.lastAckedSeq; i != c.sendSeq; i = (i + 1) % 8 {
				if _, ok := c.sentFrames[i]; !ok {
					break
				}
				delete(c.sentFrames, i)
				delete(c.sentTimes, i)
			}
		}

		switch frame.Control & 0x0F {
		case SFrameRR:
			c.isPeerReceiverReady = true
		case SFrameRNR:
			c.isPeerReceiverReady = false
		case SFrameREJ:
			// log.Printf("Received REJ for frame %d. Retransmission needed.", nr)
		case SFrameSREJ:
			if frameToResend, ok := c.sentFrames[nr]; ok {
				return EncodeFrame(frameToResend.DA, frameToResend.SA, frameToResend.Control, frameToResend.Information, frameToResend.Segmented)
			}
		}
	case UFrameFRMR:
		// Receiving a Frame Reject is a fatal error for the connection
		c.state = StateDisconnected
		return nil, ErrFrameRejected
	default:
		// Unhandled frame type, respond with FRMR
		frmrInfo := []byte{frame.Control}
		frmrFrame := &HDLCFrame{DA: frame.SA, SA: frame.DA, Type: FrameTypeU, Control: UFrameFRMR, Information: frmrInfo}
		return EncodeFrame(frmrFrame.DA, frmrFrame.SA, frmrFrame.Control, frmrFrame.Information, frmrFrame.Segmented)
	}
	return nil, nil
}

// SendData generates one or more I-frames for the given data payload, handling segmentation if necessary.
func (c *HDLCConnection) SendData(data []byte) ([][]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.state != StateConnected {
		return nil, ErrNotConnected
	}

	if (c.sendSeq-c.lastAckedSeq)%8 >= uint8(c.windowSize) {
		return nil, &HDLCError{Message: "sending window is full"}
	}

	if !c.isPeerReceiverReady {
		return nil, &HDLCError{Message: "peer receiver is not ready (RNR)"}
	}

	var frames [][]byte
	remainingData := data
	isSegmented := len(data) > c.maxFrameSize

	for len(remainingData) > 0 {
		chunkSize := len(remainingData)
		if chunkSize > c.maxFrameSize {
			chunkSize = c.maxFrameSize
		}
		chunk := remainingData[:chunkSize]
		remainingData = remainingData[chunkSize:]

		isLastSegment := len(remainingData) == 0

		frame := &HDLCFrame{
			DA:          c.destAddr,
			SA:          c.srcAddr,
			Type:        FrameTypeI,
			NS:          c.sendSeq,
			NR:          c.recvSeq,
			Information: chunk,
			Segmented:   isSegmented && !isLastSegment,
		}
		frame.Control = (frame.NS << 1) | (frame.NR << 5)

		if isLastSegment {
			frame.PF = true
		}

		encodedFrame, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, frame.Segmented)
		if err != nil {
			return nil, err
		}
		frames = append(frames, encodedFrame)

		c.sentFrames[frame.NS] = frame
		c.sentTimes[frame.NS] = time.Now()
		c.sendSeq = (c.sendSeq + 1) % 8
	}

	return frames, nil
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

// Handle processes an incoming byte stream, finds complete frames, and returns any response frames
func (c *HDLCConnection) Handle(data []byte) ([][]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if time.Since(c.lastActivity) > c.inactivityTimeout && c.state == StateConnected {
		c.state = StateDisconnected
		return nil, ErrInactivityTimeout
	}

	c.readBuffer.Write(data)
	var responses [][]byte

	for {
		startFlagIndex := bytes.IndexByte(c.readBuffer.Bytes(), FlagByte)
		if startFlagIndex == -1 {
			if c.readBuffer.Len() > MaxFrameSize*2 {
				c.readBuffer.Reset()
			}
			break
		}

		if startFlagIndex > 0 {
			c.readBuffer.Next(startFlagIndex)
		}

		buf := c.readBuffer.Bytes()
		if len(buf) < 3 {
			break
		}

		format := binary.BigEndian.Uint16(buf[1:3])
		if (format>>12)&0xF != 0xA {
			c.readBuffer.Next(1)
			continue
		}
		length := int(format & 0x7FF)

		totalFrameSize := 1 + (2 + length) + 2 + 1
		if len(buf) < totalFrameSize {
			break
		}

		frameData := buf[:totalFrameSize]
		if frameData[len(frameData)-1] != FlagByte {
			c.readBuffer.Next(1)
			continue
		}

		frameBody := frameData[1 : len(frameData)-1]
		decodedFrame, err := DecodeFrame(frameBody)
		if err == nil {
			response, err := c.HandleFrame(decodedFrame)
			if err == nil && response != nil {
				responses = append(responses, response)
			}
		}

		c.readBuffer.Next(totalFrameSize)
	}

	return responses, nil
}

// Read blocks until a complete PDU has been reassembled or a timeout occurs.
func (c *HDLCConnection) Read() ([]byte, error) {
	select {
	case pdu := <-c.ReassembledData:
		return pdu, nil
	case <-time.After(c.inactivityTimeout):
		return nil, &HDLCError{Message: "read timeout"}
	}
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
