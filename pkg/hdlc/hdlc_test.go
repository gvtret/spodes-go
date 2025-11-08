package hdlc

import (
	"bytes"
	"testing"
	"time"
)

// TestFullConnectionLifecycle validates the complete HDLC connection flow
func TestFullConnectionLifecycle(t *testing.T) {
	client := NewHDLCConnection(nil)
	server := NewHDLCConnection(nil)

	da, sa := []byte{0x01}, []byte{0x02}
	client.SetAddress(da, sa)
	server.SetAddress(sa, da)

	snrmFrameBytes, err := client.Connect()
	if err != nil {
		t.Fatalf("Client.Connect failed: %v", err)
	}

	snrmFrame, err := DecodeFrame(snrmFrameBytes[1 : len(snrmFrameBytes)-1])
	if err != nil {
		t.Fatalf("Failed to decode SNRM: %v", err)
	}
	uaFrameBytes, err := server.HandleFrame(snrmFrame)
	if err != nil {
		t.Fatalf("Server.HandleFrame(SNRM) failed: %v", err)
	}

	uaFrame, err := DecodeFrame(uaFrameBytes[1 : len(uaFrameBytes)-1])
	if err != nil {
		t.Fatalf("Failed to decode UA: %v", err)
	}
	_, err = client.HandleFrame(uaFrame)
	if err != nil {
		t.Fatalf("Client.HandleFrame(UA) failed: %v", err)
	}

	if !client.IsConnected() {
		t.Fatalf("Client should be connected")
	}

	discFrameBytes, err := client.Disconnect()
	if err != nil {
		t.Fatalf("Client.Disconnect failed: %v", err)
	}

	discFrame, err := DecodeFrame(discFrameBytes[1 : len(discFrameBytes)-1])
	if err != nil {
		t.Fatalf("Failed to decode DISC: %v", err)
	}
	uaFrameBytes, err = server.HandleFrame(discFrame)
	if err != nil {
		t.Fatalf("Server.HandleFrame(DISC) failed: %v", err)
	}

	uaFrame, err = DecodeFrame(uaFrameBytes[1 : len(uaFrameBytes)-1])
	if err != nil {
		t.Fatalf("Failed to decode UA for DISC: %v", err)
	}
	_, err = client.HandleFrame(uaFrame)
	if err != nil {
		t.Fatalf("Client.HandleFrame(UA for DISC) failed: %v", err)
	}

	if client.IsConnected() {
		t.Fatalf("Client should be disconnected")
	}
}

// TestSegmentationAndReassembly validates sending and receiving a segmented PDU
func TestSegmentationAndReassembly(t *testing.T) {
	config := DefaultConfig()
	config.MaxFrameSize = 32
	client := NewHDLCConnection(config)
	server := NewHDLCConnection(nil)

	client.SetAddress([]byte{0x11}, []byte{0x22})
	server.SetAddress([]byte{0x22}, []byte{0x11})
	client.state = StateConnected
	server.state = StateConnected

	longPDU := bytes.Repeat([]byte("s"), client.maxFrameSize*3+10)

	frames, err := client.SendData(longPDU)
	if err != nil {
		t.Fatalf("Client.SendData for segmented PDU failed: %v", err)
	}

	for _, frameBytes := range frames {
		frame, err := DecodeFrame(frameBytes[1 : len(frameBytes)-1])
		if err != nil {
			t.Fatalf("Failed to decode frame segment: %v", err)
		}
		_, err = server.HandleFrame(frame)
		if err != nil {
			t.Fatalf("Server.HandleFrame for a segment failed: %v", err)
		}
	}

	select {
	case reassembledPDU := <-server.ReassembledData:
		if !bytes.Equal(longPDU, reassembledPDU) {
			t.Errorf("Reassembled PDU does not match original PDU")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Server did not reassemble the PDU in time")
	}
}

// TestSlidingWindow validates that the sender stops when the window is full
func TestSlidingWindow(t *testing.T) {
	config := DefaultConfig()
	config.WindowSize = 2
	client := NewHDLCConnection(config)
	client.SetAddress([]byte{0x33}, []byte{0x44})
	client.state = StateConnected

	for i := 0; i < client.windowSize; i++ {
		_, err := client.SendData([]byte{byte(i)})
		if err != nil {
			t.Fatalf("SendData should not have failed yet: %v", err)
		}
	}

	_, err := client.SendData([]byte("should fail"))
	if err == nil {
		t.Fatal("SendData should have failed because the window is full")
	}
}

// TestRejectFrame simulates a lost frame and tests the REJ response
func TestRejectFrame(t *testing.T) {
	client := NewHDLCConnection(nil)
	server := NewHDLCConnection(nil)

	client.SetAddress([]byte{0x55}, []byte{0x66})
	server.SetAddress([]byte{0x66}, []byte{0x55})
	client.state = StateConnected
	server.state = StateConnected

	_, _ = client.SendData([]byte("frame1"))
	frames, _ := client.SendData([]byte("frame2"))
	frame2Bytes := frames[0]

	frame2, err := DecodeFrame(frame2Bytes[1 : len(frame2Bytes)-1])
	if err != nil {
		t.Fatalf("Failed to decode frame2: %v", err)
	}
	rejFrameBytes, err := server.HandleFrame(frame2)
	if err != nil {
		t.Fatalf("Server.HandleFrame should not fail on an out-of-order frame: %v", err)
	}

	decoded, err := DecodeFrame(rejFrameBytes[1 : len(rejFrameBytes)-1])
	if err != nil {
		t.Fatalf("Failed to decode server response: %v", err)
	}
	if decoded.Type != FrameTypeS || (decoded.Control&0x0F) != SFrameREJ {
		t.Fatal("Server should have sent a REJ frame")
	}
}

// TestReceiverNotReady validates the RNR flow control mechanism
func TestReceiverNotReady(t *testing.T) {
	client := NewHDLCConnection(nil)
	server := NewHDLCConnection(nil)

	client.SetAddress([]byte{0x77}, []byte{0x88})
	server.SetAddress([]byte{0x88}, []byte{0x77})
	client.state = StateConnected
	server.state = StateConnected

	rnrFrameBytes, _ := EncodeFrame(server.destAddr, server.srcAddr, SFrameRNR, nil, false)
	rnrFrame, err := DecodeFrame(rnrFrameBytes[1 : len(rnrFrameBytes)-1])
	if err != nil {
		t.Fatalf("Failed to decode RNR frame: %v", err)
	}

	_, err = client.HandleFrame(rnrFrame)
	if err != nil {
		t.Fatalf("Client failed to handle RNR frame: %v", err)
	}
	if client.isPeerReceiverReady {
		t.Fatal("Client should have marked the peer as not ready")
	}

	_, err = client.SendData([]byte("test"))
	if err == nil {
		t.Fatal("SendData should fail when the peer is not ready")
	}
}

// TestFrameRejectHandling verifies that the connection sends an FRMR frame for an invalid frame.
func TestFrameRejectHandling(t *testing.T) {
	server := NewHDLCConnection(nil)
	server.state = StateConnected

	// Create a deliberately invalid frame (e.g., an S-frame with an info field)
	invalidFrame := &HDLCFrame{
		DA:          []byte{0x01},
		SA:          []byte{0x02},
		Type:        FrameTypeS,
		Control:     SFrameRR,
		Information: []byte("this is not allowed"),
	}

	// This is a bit of a hack, as EncodeFrame would normally prevent this.
	// We'll manually construct a bad frame to test the server's response.
	// For this test, we'll simulate an invalid frame type instead.
	invalidFrame.Type = 99 // Not a valid frame type

	frmrResponseBytes, err := server.handleConnectedState(invalidFrame)
	if err != nil {
		t.Fatalf("handleConnectedState failed: %v", err)
	}

	frmrFrame, err := DecodeFrame(frmrResponseBytes[1 : len(frmrResponseBytes)-1])
	if err != nil {
		t.Fatalf("Failed to decode FRMR response: %v", err)
	}
	if frmrFrame.Control != UFrameFRMR {
		t.Fatal("Expected an FRMR frame in response to an invalid frame")
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
		{"FRMR-Frame", &HDLCFrame{DA: []byte{0xD}, SA: []byte{0xE}, Type: FrameTypeU, Control: UFrameFRMR, Information: []byte{0x01}}},
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

			decoded, err := DecodeFrame(encoded[1 : len(encoded)-1])
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
