package hdlc

import (
	"bytes"
	"errors"
	"sync"
	"time"
	"net"
	"io"
)

// Connection states
const (
	StateDisconnected = "disconnected"
	StateConnecting   = "connecting"
	StateConnected    = "connected"
)

// HDLCConnection manages an HDLC connection over a Transport
type HDLCConnection struct {
	transport         Transport
	state             string
	sendSeq           uint8
	recvSeq           uint8
	windowSize        int
	sentFrames        map[uint8]*HDLCFrame
	recvBuffer        map[uint8]*HDLCFrame
	mutex             sync.Mutex
	interOctetTimeout time.Duration
	inactivityTimeout time.Duration
	lastActivity      time.Time
}

// NewHDLCConnection creates a new HDLCConnection with specified transport
func NewHDLCConnection(transport Transport) *HDLCConnection {
	return &HDLCConnection{
		transport:         transport,
		state:             StateDisconnected,
		sendSeq:           0,
		recvSeq:           0,
		windowSize:        MaxWindowSize,
		sentFrames:        make(map[uint8]*HDLCFrame),
		recvBuffer:        make(map[uint8]*HDLCFrame),
		interOctetTimeout: time.Duration(InterOctetTimeout) * time.Millisecond,
		inactivityTimeout: time.Duration(InactivityTimeout) * time.Millisecond,
		lastActivity:      time.Now(),
	}
}

// Connect establishes the HDLC connection using SNRM/UA exchange
func (c *HDLCConnection) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state != StateDisconnected {
		return errors.New("already connected or connecting")
	}
	c.state = StateConnecting
	snrm := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: UFrameSNRM,
		Type:    FrameTypeU,
		PF:      true,
	}
	data, err := EncodeFrame(snrm.DA, snrm.SA, snrm.Control, nil, false)
	if err != nil {
		return err
	}
	_, err = c.transport.Write(data)
	if err != nil {
		return err
	}
	c.lastActivity = time.Now()
	frame, err := c.readFrame()
	if err != nil {
		return err
	}
	if frame.Type == FrameTypeU && frame.Control == UFrameUA {
		c.state = StateConnected
		return nil
	}
	return errors.New("unexpected response to SNRM")
}

// SendData sends data, segmenting if necessary
func (c *HDLCConnection) SendData(data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state != StateConnected {
		return errors.New("not connected")
	}
	segments := segmentData(data, MaxFrameSize)
	for i, segment := range segments {
		if len(c.sentFrames) >= c.windowSize {
			if err := c.waitForAck(); err != nil {
				return err
			}
		}
		segmented := i < len(segments)-1
		frame := &HDLCFrame{
			DA:          []byte{0xFF},
			SA:          []byte{0x01},
			Control:     (c.sendSeq << 1) | (uint8(c.recvSeq) << 5),
			Information: segment,
			Type:        FrameTypeI,
			NS:          c.sendSeq,
			NR:          c.recvSeq,
			Segmented:   segmented,
		}
		encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, segmented)
		if err != nil {
			return err
		}
		_, err = c.transport.Write(encoded)
		if err != nil {
			return err
		}
		c.sentFrames[c.sendSeq] = frame
		c.sendSeq = (c.sendSeq + 1) % 8
		c.lastActivity = time.Now()
	}
	return nil
}

// segmentData splits data into segments
func segmentData(data []byte, maxSize int) [][]byte {
	var segments [][]byte
	for len(data) > 0 {
		size := len(data)
		if size > maxSize {
			size = maxSize
		}
		segments = append(segments, data[:size])
		data = data[size:]
	}
	return segments
}

// waitForAck waits for acknowledgment or reject frames
func (c *HDLCConnection) waitForAck() error {
	for {
		frame, err := c.readFrame()
		if err != nil {
			return err
		}
		if frame.Type == FrameTypeS {
			if frame.Control&0x0F == SFrameRR {
				c.handleAck(frame.NR)
				return nil
			} else if frame.Control&0x0F == SFrameREJ {
				return c.handleReject(frame.NR)
			}
		}
	}
}

// handleAck processes acknowledgment
func (c *HDLCConnection) handleAck(nr uint8) {
	for seq := range c.sentFrames {
		if seq < nr {
			delete(c.sentFrames, seq)
		}
	}
	c.recvSeq = nr
}

// handleReject processes reject request by resending frames
func (c *HDLCConnection) handleReject(nr uint8) error {
	for seq := nr; seq != c.sendSeq; seq = (seq + 1) % 8 {
		if frame, exists := c.sentFrames[seq]; exists {
			encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, frame.Segmented)
			if err != nil {
				return err
			}
			_, err = c.transport.Write(encoded)
			if err != nil {
				return err
			}
			c.lastActivity = time.Now()
		}
	}
	return nil
}

// ReceiveData receives and assembles data from I-frames
func (c *HDLCConnection) ReceiveData() ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state != StateConnected {
		return nil, errors.New("not connected")
	}
	var result []byte
	expectedSeq := c.recvSeq

	for {
		frame, err := c.readFrame()
		if err != nil {
			return nil, err
		}
		if frame.Type != FrameTypeI {
			continue
		}

		if frame.NS != expectedSeq {
			rej := &HDLCFrame{
				DA:      frame.SA,
				SA:      frame.DA,
				Control: (expectedSeq << 5) | SFrameREJ,
				Type:    FrameTypeS,
				NR:      expectedSeq,
			}
			encoded, err := EncodeFrame(rej.DA, rej.SA, rej.Control, nil, false)
			if err != nil {
				return nil, err
			}
			_, err = c.transport.Write(encoded)
			if err != nil {
				return nil, err
			}
			continue
		}

		result = append(result, frame.Information...)
		expectedSeq = (frame.NS + 1) % 8
		c.recvSeq = expectedSeq

		rr := &HDLCFrame{
			DA:      frame.SA,
			SA:      frame.DA,
			Control: (expectedSeq << 5) | SFrameRR,
			Type:    FrameTypeS,
			NR:      expectedSeq,
		}
		encoded, err := EncodeFrame(rr.DA, rr.SA, rr.Control, nil, false)
		if err != nil {
			return nil, err
		}
		_, err = c.transport.Write(encoded)
		if err != nil {
			return nil, err
		}

		if !frame.Segmented {
			return result, nil
		}
	}
}

// Disconnect terminates the HDLC connection with DISC/UA
func (c *HDLCConnection) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state != StateConnected {
		return errors.New("not connected")
	}
	disc := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: UFrameDISC,
		Type:    FrameTypeU,
		PF:      true,
	}
	data, err := EncodeFrame(disc.DA, disc.SA, disc.Control, nil, false)
	if err != nil {
		return err
	}
	_, err = c.transport.Write(data)
	if err != nil {
		return err
	}
	c.state = StateDisconnected
	return c.transport.Close()
}

// readFrame reads a single HDLC frame with inter-octet timeout
func (c *HDLCConnection) readFrame() (*HDLCFrame, error) {
	var buffer bytes.Buffer
	for {
		if time.Since(c.lastActivity) > c.inactivityTimeout {
			c.state = StateDisconnected
			c.transport.Close()
			return nil, errors.New("inactivity timeout")
		}
		b := make([]byte, 1)
		err := c.transport.SetReadDeadline(time.Now().Add(c.interOctetTimeout))
		if err != nil {
			return nil, err
		}
		n, err := c.transport.Read(b)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil, err
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if buffer.Len() > 0 {
					return nil, errors.New("inter-octet timeout")
				}
				continue
			}
			return nil, err
		}
		if n == 0 {
			continue
		}
		c.lastActivity = time.Now()
		buffer.Write(b)
		if b[0] == FlagByte && buffer.Len() > 1 {
			frameData := buffer.Bytes()
			if frameData[0] == FlagByte {
				frame, err := DecodeFrame(frameData, c.interOctetTimeout)
				if err != nil {
					if err.Error() == "incomplete frame received" {
						continue
					}
					return nil, err
				}
				return frame, nil
			}
			buffer.Reset()
			buffer.WriteByte(FlagByte)
		}
	}
}

// Handle processes incoming HDLC frames and manages connection state
func (c *HDLCConnection) Handle() error {
	for {
		frame, err := c.readFrame()
		if err != nil {
			return err
		}
		c.mutex.Lock()
		switch c.state {
		case StateDisconnected:
			if frame.Type == FrameTypeU && frame.Control == UFrameSNRM {
				uaFrame := &HDLCFrame{
					Type:    FrameTypeU,
					Control: UFrameUA,
					DA:      frame.SA,
					SA:      frame.DA,
				}
				encoded, err := EncodeFrame(uaFrame.DA, uaFrame.SA, uaFrame.Control, nil, false)
				if err != nil {
					c.mutex.Unlock()
					return err
				}
				_, err = c.transport.Write(encoded)
				if err != nil {
					c.mutex.Unlock()
					return err
				}
				c.state = StateConnected
				c.lastActivity = time.Now()
			}
		case StateConnected:
			if frame.Type == FrameTypeI {
				// Validate both N(S) and N(R)
				if frame.NS != c.recvSeq || frame.NR != c.sendSeq {
					rejFrame := &HDLCFrame{
						Type:    FrameTypeS,
						Control: SFrameREJ | (c.recvSeq << 5),
						DA:      frame.SA,
						SA:      frame.DA,
					}
					encoded, err := EncodeFrame(rejFrame.DA, rejFrame.SA, rejFrame.Control, nil, false)
					if err != nil {
						c.mutex.Unlock()
						return err
					}
					_, err = c.transport.Write(encoded)
					if err != nil {
						c.mutex.Unlock()
						return err
					}
					c.mutex.Unlock()
					continue
				}
				data := frame.Information
				c.recvSeq = (c.recvSeq + 1) % 8
				rrFrame := &HDLCFrame{
					Type:    FrameTypeS,
					Control: SFrameRR | (c.recvSeq << 5),
					DA:      frame.SA,
					SA:      frame.DA,
				}
				encoded, err := EncodeFrame(rrFrame.DA, rrFrame.SA, rrFrame.Control, nil, false)
				if err != nil {
					c.mutex.Unlock()
					return err
				}
				_, err = c.transport.Write(encoded)
				if err != nil {
					c.mutex.Unlock()
					return err
				}
				responseFrame := &HDLCFrame{
					Type:        FrameTypeI,
					Control:     (c.sendSeq << 1) | (c.recvSeq << 5),
					Information: data,
					DA:          frame.SA,
					SA:          frame.DA,
					Segmented:   frame.Segmented,
				}
				encoded, err = EncodeFrame(responseFrame.DA, responseFrame.SA, responseFrame.Control, responseFrame.Information, frame.Segmented)
				if err != nil {
					c.mutex.Unlock()
					return err
				}
				_, err = c.transport.Write(encoded)
				if err != nil {
					c.mutex.Unlock()
					return err
				}
				c.sentFrames[c.sendSeq] = responseFrame
				c.sendSeq = (c.sendSeq + 1) % 8
				c.lastActivity = time.Now()
			} else if frame.Type == FrameTypeU && frame.Control == UFrameDISC {
				uaFrame := &HDLCFrame{
					Type:    FrameTypeU,
					Control: UFrameUA,
					DA:      frame.SA,
					SA:      frame.DA,
				}
				encoded, err := EncodeFrame(uaFrame.DA, uaFrame.SA, uaFrame.Control, nil, false)
				if err != nil {
					c.mutex.Unlock()
					return err
				}
				_, err = c.transport.Write(encoded)
				if err != nil {
					c.mutex.Unlock()
					return err
				}
				c.state = StateDisconnected
				c.mutex.Unlock()
				return c.transport.Close()
			}
		}
		c.mutex.Unlock()
	}
}