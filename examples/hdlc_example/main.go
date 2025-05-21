package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gvtret/spodes-go/pkg/hdlc"
)

func main() {
	// 1. Define parameters for an HDLC UI frame
	destAddress := []byte{0x01, 0x02} // Example 2-byte destination address
	sourceAddress := []byte{0x03, 0x04} // Example 2-byte source address
	// For a UI frame, the control byte is simply hdlc.UFrameUI
	controlByte := byte(hdlc.UFrameUI)
	informationPayload := []byte("hello world")

	fmt.Println("--- HDLC Frame Encoding & Decoding Example ---")
	fmt.Printf("Original Frame Parameters:\n")
	fmt.Printf("  Destination Address: %X\n", destAddress)
	fmt.Printf("  Source Address: %X\n", sourceAddress)
	fmt.Printf("  Control Byte: 0x%02X (UI Frame)\n", controlByte)
	fmt.Printf("  Information Payload: \"%s\"\n\n", string(informationPayload))

	// 2. Encode the HDLC UI frame
	// The 'segmented' flag is false for a single UI frame.
	encodedFrameBytes, err := hdlc.EncodeFrame(destAddress, sourceAddress, controlByte, informationPayload, false)
	if err != nil {
		log.Fatalf("Error encoding HDLC frame: %v", err)
	}
	fmt.Printf("Encoded HDLC Frame Bytes: %X\n\n", encodedFrameBytes)

	// 3. Decode the HDLC frame
	// The interOctetTimeout is relevant for stream-based decoding, for a complete frame buffer,
	// it acts more like a general processing timeout if internal logic had such delays.
	// For decoding a byte buffer that is already complete, a short timeout is fine.
	decodedFrame, err := hdlc.DecodeFrame(encodedFrameBytes, 200*time.Millisecond)
	if err != nil {
		log.Fatalf("Error decoding HDLC frame: %v", err)
	}

	fmt.Printf("Decoded HDLC Frame Fields:\n")
	fmt.Printf("  Frame Type Indicator: 0x%X (Format Field)\n", decodedFrame.Format) // Example, actual format field
	fmt.Printf("  Destination Address: %X\n", decodedFrame.DA)
	fmt.Printf("  Source Address: %X\n", decodedFrame.SA)
	fmt.Printf("  Control Byte: 0x%02X\n", decodedFrame.Control)
	fmt.Printf("  Frame Type (parsed): %d (0=I, 1=S, 2=U)\n", decodedFrame.Type)
	if decodedFrame.HCS != 0 { // HCS is present for frames with information
		fmt.Printf("  Header Check Sequence (HCS): 0x%04X\n", decodedFrame.HCS)
	}
	fmt.Printf("  Information Payload: \"%s\"\n", string(decodedFrame.Information))
	fmt.Printf("  Frame Check Sequence (FCS): 0x%04X\n\n", decodedFrame.FCS) // Note: FCS is validated internally

	// 4. Verification (simple check)
	if string(decodedFrame.Information) == string(informationPayload) &&
		equalBytes(decodedFrame.DA, destAddress) &&
		equalBytes(decodedFrame.SA, sourceAddress) &&
		decodedFrame.Control == controlByte {
		fmt.Println("Successfully encoded and decoded HDLC UI frame. Original and recovered data match.")
	} else {
		fmt.Println("Data mismatch between original parameters and decoded frame.")
	}
}

// Helper function to compare byte slices
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
