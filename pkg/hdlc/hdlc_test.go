package hdlc

import (
	"bytes"
	"testing"
)

// TestFullConnectionLifecycle validates the complete HDLC connection flow
func TestFullConnectionLifecycle(t *testing.T) {
	client := NewHDLCConnection()
	server := NewHDLCConnection()

	// Set addresses
	da, sa := []byte{0x01}, []byte{0x02}
	client.SetAddress(da, sa)
	server.SetAddress(sa, da)

	// Client sends SNRM
	snrmFrame, err := client.Connect()
	if err != nil {
		t.Fatalf("Client.Connect failed: %v", err)
	}

	// Server handles SNRM and responds with UA
	uaFrame, err := server.HandleFrame(snrmFrame)
	if err != nil {
		t.Fatalf("Server.HandleFrame(SNRM) failed: %v", err)
	}

	// Client handles UA
	_, err = client.HandleFrame(uaFrame)
	if err != nil {
		t.Fatalf("Client.HandleFrame(UA) failed: %v", err)
	}

	if !client.IsConnected() {
		t.Fatalf("Client should be connected")
	}

	// Client sends DISC
	discFrame, err := client.Disconnect()
	if err != nil {
		t.Fatalf("Client.Disconnect failed: %v", err)
	}

	// Server handles DISC and responds with UA
	uaFrame, err = server.HandleFrame(discFrame)
	if err != nil {
		t.Fatalf("Server.HandleFrame(DISC) failed: %v", err)
	}

	// Client handles UA
	_, err = client.HandleFrame(uaFrame)
	if err != nil {
		t.Fatalf("Client.HandleFrame(UA) failed: %v", err)
	}

	if client.IsConnected() {
		t.Fatalf("Client should be disconnected")
	}
}

// TestDataTransfer validates sending and receiving I-frames
func TestDataTransfer(t *testing.T) {
	client := NewHDLCConnection()
	server := NewHDLCConnection()

	da, sa := []byte{0x03}, []byte{0x04}
	client.SetAddress(da, sa)
	server.SetAddress(sa, da)

	// Manually set both to connected state for this test
	client.state = StateConnected
	server.state = StateConnected

	// Client sends I-frame
	testData := []byte("test data")
	iFrame, err := client.SendData(testData)
	if err != nil {
		t.Fatalf("Client.SendData failed: %v", err)
	}

	// Server handles I-frame and should respond with RR
	rrFrame, err := server.HandleFrame(iFrame)
	if err != nil {
		t.Fatalf("Server.HandleFrame(I-frame) failed: %v", err)
	}

	// Verify server sent an RR frame
	decodedRR, err := DecodeFrame(rrFrame, 0)
	if err != nil {
		t.Fatalf("Failed to decode RR frame: %v", err)
	}
	if decodedRR.Type != FrameTypeS || (decodedRR.Control&0x0F) != SFrameRR {
		t.Errorf("Expected RR frame, but got something else")
	}
}

// TestFrameRejection validates REJ frame handling
func TestFrameRejection(t *testing.T) {
	client := NewHDLCConnection()
	server := NewHDLCConnection()

	da, sa := []byte{0x05}, []byte{0x06}
	client.SetAddress(da, sa)
	server.SetAddress(sa, da)

	client.state = StateConnected
	server.state = StateConnected

	// Manually advance client's send sequence to create an out-of-order frame
	client.sendSeq = 1

	iFrame, _ := client.SendData([]byte("out of order"))

	// Server should detect the sequence error and respond with REJ
	// Note: The current HandleFrame logic does not implement REJ, so this will fail
	// This test is written to guide the implementation of that feature.
	rejFrame, err := server.HandleFrame(iFrame)
	if err != nil {
		t.Fatalf("Server.HandleFrame failed unexpectedly: %v", err)
	}

	decodedREJ, err := DecodeFrame(rejFrame, 0)
	if err != nil {
		t.Fatalf("Failed to decode REJ frame: %v", err)
	}
	if decodedREJ.Type != FrameTypeS || (decodedREJ.Control&0x0F) != SFrameREJ {
		t.Errorf("Expected REJ frame, but got something else")
	}
}

// TestFrameEncodeDecode ensures all frame types are correctly handled
func TestFrameEncodeDecode(t *testing.T) {
	testCases := []struct {
		name  string
		frame *HDLCFrame
	}{
		{"I-Frame", &HDLCFrame{DA: []byte{0x1}, SA: []byte{0x2}, Type: FrameTypeI, NS: 1, NR: 2, Information: []byte("test")}},
		{"RR-Frame", &HDLCFrame{DA: []byte{0x3}, SA: []byte{0x4}, Type: FrameTypeS, Control: SFrameRR | (3 << 5)}},
		{"REJ-Frame", &HDLCFrame{DA: []byte{0x5}, SA: []byte{0x6}, Type: FrameTypeS, Control: SFrameREJ | (4 << 5)}},
		{"SNRM-Frame", &HDLCFrame{DA: []byte{0x7}, SA: []byte{0x8}, Type: FrameTypeU, Control: UFrameSNRM}},
		{"UA-Frame", &HDLCFrame{DA: []byte{0x9}, SA: []byte{0xA}, Type: FrameTypeU, Control: UFrameUA}},
		{"DISC-Frame", &HDLCFrame{DA: []byte{0xB}, SA: []byte{0xC}, Type: FrameTypeU, Control: UFrameDISC}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.frame.Type == FrameTypeI {
				tc.frame.Control = (tc.frame.NS << 1) | (tc.frame.NR << 5)
			}
			encoded, err := EncodeFrame(tc.frame.DA, tc.frame.SA, tc.frame.Control, tc.frame.Information, false)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}
			decoded, err := DecodeFrame(encoded, 0)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}
			if !bytes.Equal(decoded.DA, tc.frame.DA) || !bytes.Equal(decoded.SA, tc.frame.SA) {
				t.Errorf("Address mismatch")
			}
			if decoded.Control != tc.frame.Control {
				t.Errorf("Control field mismatch")
			}
			if !bytes.Equal(decoded.Information, tc.frame.Information) {
				t.Errorf("Information mismatch")
			}
		})
	}
}
