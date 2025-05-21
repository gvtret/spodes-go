package hdlc

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"
)

const host_address = "localhost:8080"

// Error message constants for test assertions
const (
	errEncodeFailed      = "Encode failed: %v"
	errDecodeFailed      = "Decode failed: %v"
	errListenFailed      = "Listen failed: %v"
	errDialFailed        = "Dial Failed: %v"
	errConnectFailed     = "Connect failed: %v"
	errWriteFailed       = "Write failed: %v"
	errReadFrameFailed   = "Read frame failed: %v"
	errSendDataFailed    = "SendData failed: %v"
	errReceiveDataFailed = "ReceiveData failed: %v"
	errDisconnectFailed  = "Disconnect failed: %v"
)

// TestFrameEncodeDecodeI tests encoding and decoding of I-frame
func TestFrameEncodeDecodeI(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{0xFF},
		SA:          []byte{0x01},
		Control:     0x00, // I-frame, N(S)=0, N(R)=0
		Information: []byte("test"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || !bytes.Equal(decoded.Information, frame.Information) || decoded.Control != frame.Control {
		t.Errorf("Decoded I-frame mismatch: got DA=%v, SA=%v, Info=%v, Control=0x%X; want DA=%v, SA=%v, Info=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Information, decoded.Control, frame.DA, frame.SA, frame.Information, frame.Control)
	}
	if decoded.Type != FrameTypeI || decoded.NS != 0 || decoded.NR != 0 {
		t.Errorf("I-frame fields mismatch: got Type=%d, NS=%d, NR=%d; want Type=%d, NS=0, NR=0", decoded.Type, decoded.NS, decoded.NR, FrameTypeI)
	}
}

// TestFrameEncodeDecodeSRR tests encoding and decoding of S-frame (RR)
func TestFrameEncodeDecodeSRR(t *testing.T) {
	frame := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: SFrameRR | (2 << 5), // S-frame RR, N(R)=2
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, nil, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded S-frame (RR) mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
	if decoded.Type != FrameTypeS || decoded.NR != 2 || len(decoded.Information) != 0 {
		t.Errorf("S-frame (RR) fields mismatch: got Type=%d, NR=%d, Info=%v; want Type=%d, NR=2, Info=[]", decoded.Type, decoded.NR, decoded.Information, FrameTypeS)
	}
}

// TestFrameEncodeDecodeSRNR tests encoding and decoding of S-frame (RNR)
func TestFrameEncodeDecodeSRNR(t *testing.T) {
	frame := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: SFrameRNR | (3 << 5), // S-frame RNR, N(R)=3
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, nil, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded S-frame (RNR) mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
	if decoded.Type != FrameTypeS || decoded.NR != 3 || len(decoded.Information) != 0 {
		t.Errorf("S-frame (RNR) fields mismatch: got Type=%d, NR=%d, Info=%v; want Type=%d, NR=3, Info=[]", decoded.Type, decoded.NR, decoded.Information, FrameTypeS)
	}
}

// TestFrameEncodeDecodeSREJ tests encoding and decoding of S-frame (REJ)
func TestFrameEncodeDecodeSREJ(t *testing.T) {
	frame := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: SFrameREJ | (4 << 5), // S-frame REJ, N(R)=4
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, nil, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded S-frame (REJ) mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
	if decoded.Type != FrameTypeS || decoded.NR != 4 || len(decoded.Information) != 0 {
		t.Errorf("S-frame (REJ) fields mismatch: got Type=%d, NR=%d, Info=%v; want Type=%d, NR=4, Info=[]", decoded.Type, decoded.NR, decoded.Information, FrameTypeS)
	}
}

// TestFrameEncodeDecodeUSNRM tests encoding and decoding of U-frame (SNRM)
func TestFrameEncodeDecodeUSNRM(t *testing.T) {
	frame := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: UFrameSNRM,
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, nil, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded U-frame (SNRM) mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
	if decoded.Type != FrameTypeU || len(decoded.Information) != 0 {
		t.Errorf("U-frame (SNRM) fields mismatch: got Type=%d, Info=%v; want Type=%d, Info=[]", decoded.Type, decoded.Information, FrameTypeU)
	}
}

// TestFrameEncodeDecodeUUA tests encoding and decoding of U-frame (UA)
func TestFrameEncodeDecodeUUA(t *testing.T) {
	frame := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: UFrameUA,
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, nil, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded U-frame (UA) mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
	if decoded.Type != FrameTypeU || len(decoded.Information) != 0 {
		t.Errorf("U-frame (UA) fields mismatch: got Type=%d, Info=%v; want Type=%d, Info=[]", decoded.Type, decoded.Information, FrameTypeU)
	}
}

// TestFrameEncodeDecodeUDISC tests encoding and decoding of U-frame (DISC)
func TestFrameEncodeDecodeUDISC(t *testing.T) {
	frame := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: UFrameDISC,
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, nil, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded U-frame (DISC) mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
	if decoded.Type != FrameTypeU || len(decoded.Information) != 0 {
		t.Errorf("U-frame (DISC) fields mismatch: got Type=%d, Info=%v; want Type=%d, Info=[]", decoded.Type, decoded.Information, FrameTypeU)
	}
}

// TestFrameEncodeDecodeUDM tests encoding and decoding of U-frame (DM)
func TestFrameEncodeDecodeUDM(t *testing.T) {
	frame := &HDLCFrame{
		DA:      []byte{0xFF},
		SA:      []byte{0x01},
		Control: UFrameDM,
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, nil, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded U-frame (DM) mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
	if decoded.Type != FrameTypeU || len(decoded.Information) != 0 {
		t.Errorf("U-frame (DM) fields mismatch: got Type=%d, Info=%v; want Type=%d, Info=[]", decoded.Type, decoded.Information, FrameTypeU)
	}
}

// TestFrameEncodeDecodeUFRMR tests encoding and decoding of U-frame (FRMR)
func TestFrameEncodeDecodeUFRMR(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{0xFF},
		SA:          []byte{0x01},
		Control:     UFrameFRMR,
		Information: []byte{0x00, 0x00, 0x00}, // FRMR data (example)
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control || !bytes.Equal(decoded.Information, frame.Information) {
		t.Errorf("Decoded U-frame (FRMR) mismatch: got DA=%v, SA=%v, Control=0x%X, Info=%v; want DA=%v, SA=%v, Control=0x%X, Info=%v",
			decoded.DA, decoded.SA, decoded.Control, decoded.Information, frame.DA, frame.SA, frame.Control, frame.Information)
	}
	if decoded.Type != FrameTypeU {
		t.Errorf("U-frame (FRMR) fields mismatch: got Type=%d; want Type=%d", decoded.Type, FrameTypeU)
	}
}

// TestFrameEncodeDecodeUUI tests encoding and decoding of U-frame (UI)
func TestFrameEncodeDecodeUUI(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{0xFF},
		SA:          []byte{0x01},
		Control:     UFrameUI,
		Information: []byte("unconfirmed"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control || !bytes.Equal(decoded.Information, frame.Information) {
		t.Errorf("Decoded U-frame (UI) mismatch: got DA=%v, SA=%v, Control=0x%X, Info=%v; want DA=%v, SA=%v, Control=0x%X, Info=%v",
			decoded.DA, decoded.SA, decoded.Control, decoded.Information, frame.DA, frame.SA, frame.Control, frame.Information)
	}
	if decoded.Type != FrameTypeU {
		t.Errorf("U-frame (UI) fields mismatch: got Type=%d; want Type=%d", decoded.Type, FrameTypeU)
	}
}

// TestFrameEncodeDecodeAddress1Byte tests encoding and decoding with 1-byte server address
func TestFrameEncodeDecodeAddress1Byte(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{0x03}, // 1-byte logical server address
		SA:          []byte{0x01},
		Control:     0x00,
		Information: []byte("test"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded 1-byte address frame mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
}

// TestFrameEncodeDecodeAddress2Byte tests encoding and decoding with 2-byte server address
func TestFrameEncodeDecodeAddress2Byte(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{0x00, 0x03}, // 2-byte logical server address
		SA:          []byte{0x01},
		Control:     0x00,
		Information: []byte("test"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded 2-byte address frame mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
}

// TestFrameEncodeDecodeAddress4Byte tests encoding and decoding with 4-byte server address
func TestFrameEncodeDecodeAddress4Byte(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{0x00, 0x00, 0x00, 0x03}, // 4-byte physical server address
		SA:          []byte{0x01},
		Control:     0x00,
		Information: []byte("test"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control {
		t.Errorf("Decoded 4-byte address frame mismatch: got DA=%v, SA=%v, Control=0x%X; want DA=%v, SA=%v, Control=0x%X",
			decoded.DA, decoded.SA, decoded.Control, frame.DA, frame.SA, frame.Control)
	}
}

// TestFrameInvalidSequenceNS tests handling of incorrect N(S) sequence
func TestFrameInvalidSequenceNS(t *testing.T) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf(errListenFailed, err)
	}
	defer listener.Close()

	// Start server in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		transport := NewTCPTransport(conn)
		hdlcConn := NewHDLCConnection(transport)
		if err := hdlcConn.Handle(); err != nil {
			return
		}
	}()

	time.Sleep(100 * time.Millisecond)
	conn, err := net.Dial("tcp", host_address)
	if err != nil {
		t.Fatalf(errDialFailed, err)
	}
	defer conn.Close()
	transport := NewTCPTransport(conn)
	hdlcConn := NewHDLCConnection(transport)
	err = hdlcConn.Connect()
	if err != nil {
		t.Fatalf(errConnectFailed, err)
	}
	// Send I-frame with incorrect N(S)=2 (expected N(S)=0)
	frame := &HDLCFrame{
		DA:          []byte{0xFF},
		SA:          []byte{0x01},
		Control:     (2 << 1), // N(S)=2, N(R)=0
		Information: []byte("test"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	_, err = transport.Write(encoded)
	if err != nil {
		t.Fatalf(errWriteFailed, err)
	}
	// Expect REJ frame with N(R)=0
	receivedFrame, err := hdlcConn.readFrame()
	if err != nil {
		t.Fatalf(errReadFrameFailed, err)
	}
	if receivedFrame.Type != FrameTypeS || receivedFrame.Control&0x0F != SFrameREJ || receivedFrame.NR != 0 {
		t.Errorf("Expected REJ frame with N(R)=0, got Type=%d, Control=0x%X, NR=%d", receivedFrame.Type, receivedFrame.Control, receivedFrame.NR)
	}
	wg.Wait()
}

// TestFrameInvalidSequenceNR tests handling of incorrect N(R) sequence
func TestFrameInvalidSequenceNR(t *testing.T) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf(errListenFailed, err)
	}
	defer listener.Close()

	// Start server in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		transport := NewTCPTransport(conn)
		hdlcConn := NewHDLCConnection(transport)
		if err := hdlcConn.Handle(); err != nil {
			return
		}
	}()

	time.Sleep(100 * time.Millisecond)
	conn, err := net.Dial("tcp", host_address)
	if err != nil {
		t.Fatalf(errDialFailed, err)
	}
	defer conn.Close()
	transport := NewTCPTransport(conn)
	hdlcConn := NewHDLCConnection(transport)
	err = hdlcConn.Connect()
	if err != nil {
		t.Fatalf(errConnectFailed, err)
	}
	// Send I-frame with incorrect N(R)=3 (expected N(R)=0)
	frame := &HDLCFrame{
		DA:          []byte{0xFF},
		SA:          []byte{0x01},
		Control:     (0 << 1) | (3 << 5), // N(S)=0, N(R)=3
		Information: []byte("test"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	_, err = transport.Write(encoded)
	if err != nil {
		t.Fatalf(errWriteFailed, err)
	}
	// Expect REJ frame with N(R)=0
	receivedFrame, err := hdlcConn.readFrame()
	if err != nil {
		t.Fatalf(errReadFrameFailed, err)
	}
	if receivedFrame.Type != FrameTypeS || receivedFrame.Control&0x0F != SFrameREJ || receivedFrame.NR != 0 {
		t.Errorf("Expected REJ frame with N(R)=0, got Type=%d, Control=0x%X, NR=%d", receivedFrame.Type, receivedFrame.Control, receivedFrame.NR)
	}
	wg.Wait()
}

// TestFrameCorruptedFCS tests handling of frame with corrupted FCS
func TestFrameCorruptedFCS(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{0xFF},
		SA:          []byte{0x01},
		Control:     0x00,
		Information: []byte("test"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	// Corrupt both FCS bytes to ensure mismatch
	encoded[len(encoded)-3] ^= 0x55 // Second FCS byte
	encoded[len(encoded)-2] ^= 0x55 // First FCS byte
	_, err = DecodeFrame(encoded, time.Millisecond*200)
	if err == nil || err.Error() != "FCS mismatch" {
		t.Errorf("Expected FCS mismatch error, got: %v", err)
	}
}

// TestFrameCorruptedHCS tests handling of frame with corrupted HCS
func TestFrameCorruptedHCS(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{0xFF},
		SA:          []byte{0x01},
		Control:     0x00,
		Information: []byte("test"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	// Corrupt HCS by modifying the second HCS byte
	encoded[7] ^= 0x55 // Second HCS byte (DA=1, SA=1, control=1, HCS starts at offset 6)
	_, err = DecodeFrame(encoded, time.Millisecond*200)
	if err == nil || (err.Error() != "HCS mismatch" && err.Error() != "FCS mismatch") {
		t.Errorf("Expected HCS mismatch or FCS mismatch error, got: %v", err)
	}
}

// TestFrameBroadcastUI tests encoding and decoding of broadcast UI-frame
func TestFrameBroadcastUI(t *testing.T) {
	frame := &HDLCFrame{
		DA:          []byte{BroadcastAddress}, // Broadcast address
		SA:          []byte{0x01},
		Control:     UFrameUI,
		Information: []byte("broadcast message"),
	}
	encoded, err := EncodeFrame(frame.DA, frame.SA, frame.Control, frame.Information, false)
	if err != nil {
		t.Fatalf(errEncodeFailed, err)
	}
	decoded, err := DecodeFrame(encoded, time.Millisecond*200)
	if err != nil {
		t.Fatalf(errDecodeFailed, err)
	}
	if !bytes.Equal(decoded.DA, frame.DA) || !bytes.Equal(decoded.SA, frame.SA) || decoded.Control != frame.Control || !bytes.Equal(decoded.Information, frame.Information) {
		t.Errorf("Decoded broadcast UI-frame mismatch: got DA=%v, SA=%v, Control=0x%X, Info=%v; want DA=%v, SA=%v, Control=0x%X, Info=%v",
			decoded.DA, decoded.SA, decoded.Control, decoded.Information, frame.DA, frame.SA, frame.Control, frame.Information)
	}
	if decoded.Type != FrameTypeU {
		t.Errorf("Broadcast UI-frame fields mismatch: got Type=%d; want Type=%d", decoded.Type, FrameTypeU)
	}
}

// TestIncompleteFrame tests handling of incomplete frames
func TestIncompleteFrame(t *testing.T) {
	partialFrame := []byte{FlagByte, 0xA0, 0x07, 0xFF, 0x01, 0x00} // Missing end flag
	_, err := DecodeFrame(partialFrame, time.Millisecond*200)
	if err == nil || err.Error() != "incomplete frame received" {
		t.Errorf("Expected incomplete frame error, got: %v", err)
	}
}

// TestConnection tests full HDLC connection lifecycle over TCP
func TestConnection(t *testing.T) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf(errListenFailed, err)
	}
	defer listener.Close()

	// Start server in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		transport := NewTCPTransport(conn)
		hdlcConn := NewHDLCConnection(transport)
		if err := hdlcConn.Handle(); err != nil {
			return
		}
	}()

	time.Sleep(100 * time.Millisecond)
	conn, err := net.Dial("tcp", host_address)
	if err != nil {
		t.Fatalf(errDialFailed, err)
	}
	defer conn.Close()
	transport := NewTCPTransport(conn)
	hdlcConn := NewHDLCConnection(transport)
	err = hdlcConn.Connect()
	if err != nil {
		t.Fatalf(errConnectFailed, err)
	}
	data := []byte("hello")
	err = hdlcConn.SendData(data)
	if err != nil {
		t.Fatalf(errSendDataFailed, err)
	}
	received, err := hdlcConn.ReceiveData()
	if err != nil {
		t.Fatalf(errReceiveDataFailed, err)
	}
	if !bytes.Equal(received, data) {
		t.Errorf("Received data mismatch: got %v, want %v", received, data)
	}
	err = hdlcConn.Disconnect()
	if err != nil {
		t.Fatalf(errDisconnectFailed, err)
	}
	wg.Wait()
}

// TestInactivityTimeout tests connection reset on inactivity
func TestInactivityTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf(errListenFailed, err)
	}
	defer listener.Close()

	// Start server in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		transport := NewTCPTransport(conn)
		hdlcConn := NewHDLCConnection(transport)
		if err := hdlcConn.Handle(); err != nil {
			return
		}
	}()

	time.Sleep(100 * time.Millisecond)
	conn, err := net.Dial("tcp", host_address)
	if err != nil {
		t.Fatalf(errDialFailed, err)
	}
	defer conn.Close()
	transport := NewTCPTransport(conn)
	hdlcConn := NewHDLCConnection(transport)
	hdlcConn.inactivityTimeout = 500 * time.Millisecond
	err = hdlcConn.Connect()
	if err != nil {
		t.Fatalf(errConnectFailed, err)
	}
	time.Sleep(600 * time.Millisecond)
	_, err = hdlcConn.ReceiveData()
	if err == nil || err.Error() != "inactivity timeout" {
		t.Errorf("Expected inactivity timeout, got: %v", err)
	}
	wg.Wait()
}

// TestBitStuff tests bit stuffing and unstuffing
func TestBitStuff(t *testing.T) {
	// Test case with five consecutive ones to trigger stuffing
	input := []byte{0xF8} // 11111000 in binary
	expected := []byte{0x7C, 0x00} // 11111000 -> 111110_0_00 (stuffed zero after five ones)
	stuffed := bitStuff(input)
	if !bytes.Equal(stuffed, expected) {
		t.Errorf("BitStuff failed: got %v, want %v", stuffed, expected)
	}
	unstuffed, err := bitUnstuff(stuffed)
	if err != nil {
		t.Fatalf("BitUnstuff failed: %v", err)
	}
	if !bytes.Equal(unstuffed, input) {
		t.Errorf("BitUnstuff failed: got %v, want %v", unstuffed, input)
	}
}