package hdlc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Constants for HDLC protocol
const (
	FlagByte          = 0x7E      // Frame delimiter
	MaxWindowSize     = 7         // Maximum window size for sliding window
	MaxFrameSize      = 2048      // Maximum frame size in bytes
	InterOctetTimeout = 200       // Inter-octet timeout in milliseconds
	InactivityTimeout = 30 * 1000 // Inactivity timeout in milliseconds
	BroadcastAddress  = 0xFF      // Broadcast address for UI frames
)

// Frame types
const (
	FrameTypeI = iota // Information frame
	FrameTypeS        // Supervisory frame
	FrameTypeU        // Unnumbered frame
)

// U-frame commands
const (
	UFrameSNRM = 0x83 // Set Normal Response Mode
	UFrameUA   = 0x63 // Unnumbered Acknowledge
	UFrameDISC = 0x43 // Disconnect
	UFrameDM   = 0x0F // Disconnected Mode
	UFrameFRMR = 0x87 // Frame Reject
	UFrameUI   = 0x03 // Unnumbered Information
)

// S-frame types
const (
	SFrameRR  = 0x01 // Receive Ready
	SFrameRNR = 0x05 // Receive Not Ready
	SFrameREJ = 0x09 // Reject
)

// HDLCFrame represents an HDLC frame structure
type HDLCFrame struct {
	Format      uint16 // Frame format field (2 bytes)
	DA          []byte // Destination Address (logical or physical server address)
	SA          []byte // Source Address (client address)
	Control     byte   // Control field
	HCS         uint16 // Header Check Sequence
	Information []byte // Information field
	FCS         uint16 // Frame Check Sequence
	Type        int    // Frame type (I, S, U)
	NS          uint8  // Send sequence number
	NR          uint8  // Receive sequence number
	PF          bool   // Poll/Final bit
	Segmented   bool   // Segment flag in Format field
}

// calculateCRC16 computes the CRC-16 checksum using CCITT polynomial
func calculateCRC16(data []byte) uint16 {
	const crc16CCITT = 0x1021
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ crc16CCITT
			} else {
				crc <<= 1
			}
		}
	}
	return ^crc
}

// EncodeFrame encodes an HDLC frame according to IEC 62056-46 (no stuffing)
func EncodeFrame(da, sa []byte, control byte, info []byte, segmented bool) ([]byte, error) {
	if control&0x01 == 0 && segmented {
		control |= 0x10
	}

	encodedDA := encodeAddress(da)
	if encodedDA == nil {
		return nil, errors.New("invalid destination address")
	}
	encodedSA := encodeAddress(sa)
	if encodedSA == nil {
		return nil, errors.New("invalid source address")
	}
	hasInfo := hasInformation(control)
	if hasInfo && len(info) == 0 {
		return nil, errors.New("information field required for I or UI frame")
	} else if !hasInfo && len(info) > 0 {
		return nil, errors.New("information field not allowed for this frame type")
	}

	// The `length` field is the length of the frame *payload*, which is
	// everything between the format field and the FCS.
	payloadLength := len(encodedDA) + len(encodedSA) + 1
	if hasInfo {
		payloadLength += 2 + len(info) // HCS + Info
	}
	if payloadLength > 2047 {
		return nil, errors.New("frame too long")
	}

	// Construct the frame payload (everything after format field)
	var payload bytes.Buffer
	payload.Write(encodedDA)
	payload.Write(encodedSA)
	payload.WriteByte(control)

	// Add HCS and Information if needed
	if hasInfo {
		// HCS is calculated over Format + DA + SA + Control
		// We need to build a temporary buffer for this
		var headerForHCS bytes.Buffer
		tempFormat := uint16(0xA<<12) | uint16(payloadLength&0x7FF)
		binary.Write(&headerForHCS, binary.BigEndian, tempFormat)
		headerForHCS.Write(payload.Bytes())
		hcs := calculateCRC16(headerForHCS.Bytes())
		binary.Write(&payload, binary.BigEndian, hcs)
		payload.Write(info)
	}

	// Now build the full frame body (Format + Payload) for the FCS calculation
	var frameBody bytes.Buffer
	format := uint16(0xA<<12) | uint16(payloadLength&0x7FF)
	binary.Write(&frameBody, binary.BigEndian, format)
	frameBody.Write(payload.Bytes())

	fcs := calculateCRC16(frameBody.Bytes())

	// Construct the final frame
	var finalFrame bytes.Buffer
	finalFrame.WriteByte(FlagByte)
	finalFrame.Write(frameBody.Bytes())
	binary.Write(&finalFrame, binary.BigEndian, fcs) // Append FCS
	finalFrame.WriteByte(FlagByte)

	return finalFrame.Bytes(), nil
}

// DecodeFrame decodes a complete frame body (everything between the flags)
func DecodeFrame(frameBody []byte) (*HDLCFrame, error) {
	if len(frameBody) < 4 { // Must have at least format (2) and FCS (2)
		return nil, errors.New("frame body is too short")
	}

	payload := frameBody[:len(frameBody)-2]
	fcsReceived := binary.BigEndian.Uint16(frameBody[len(frameBody)-2:])

	fcsCalculated := calculateCRC16(payload)
	if fcsCalculated != fcsReceived {
		return nil, fmt.Errorf("FCS mismatch: received 0x%X, calculated 0x%X", fcsReceived, fcsCalculated)
	}

	format := binary.BigEndian.Uint16(payload[0:2])
	if (format>>12)&0xF != 0xA {
		return nil, errors.New("invalid format type")
	}
	length := int(format & 0x7FF)
	// The length in the format field is the length of the payload *after* the format field.
	// So, the total length of the `payload` buffer should be length + 2 bytes for the format field.
	if len(payload) != length+2 {
		return nil, fmt.Errorf("frame length mismatch: specified %d, actual %d", length, len(payload)-2)
	}

	f, err := validateFrameStructure(payload)
	if err != nil {
		return nil, err
	}

	return parseFrameControl(f, payload)
}

// validateFrameStructure works on the full payload (Format + rest)
func validateFrameStructure(payload []byte) (*HDLCFrame, error) {
	format := binary.BigEndian.Uint16(payload[0:2])
	dataPart := payload[2:]

	da, daLen := decodeAddress(dataPart)
	if daLen == 0 {
		return nil, errors.New("invalid destination address")
	}

	saStart := daLen
	if saStart >= len(dataPart) {
		return nil, errors.New("frame too short for source address")
	}
	sa, saLen := decodeAddress(dataPart[saStart:])
	if saLen == 0 {
		return nil, errors.New("invalid source address")
	}

	controlStart := saStart + saLen
	if controlStart >= len(dataPart) {
		return nil, errors.New("missing control field")
	}

	return &HDLCFrame{
		Format:  format,
		DA:      da,
		SA:      sa,
		Control: dataPart[controlStart],
	}, nil
}

// parseFrameControl works on the full payload (Format + rest)
func parseFrameControl(f *HDLCFrame, payload []byte) (*HDLCFrame, error) {
	dataPart := payload[2:]
	controlStart := len(encodeAddress(f.DA)) + len(encodeAddress(f.SA))
	hasInfo := hasInformation(f.Control)

	if hasInfo {
		hcsStart := controlStart + 1
		if hcsStart+2 > len(dataPart) {
			return nil, errors.New("missing HCS")
		}
		f.HCS = binary.BigEndian.Uint16(dataPart[hcsStart : hcsStart+2])

		// HCS is calculated over Format + DA + SA + Control
		headerForHCS := payload[:2+controlStart+1]
		if calculateCRC16(headerForHCS) != f.HCS {
			return nil, errors.New("HCS mismatch")
		}

		f.Information = dataPart[hcsStart+2:]
	} else {
		if controlStart+1 != len(dataPart) {
			return nil, errors.New("unexpected data after control field in non-info frame")
		}
	}

	if f.Control&0x01 == 0 {
		f.Type = FrameTypeI
		f.NS = (f.Control >> 1) & 0x07
		f.PF = (f.Control & 0x10) != 0
		f.NR = (f.Control >> 5) & 0x07
		f.Segmented = f.PF
	} else if f.Control&0x03 == 0x01 {
		f.Type = FrameTypeS
		f.PF = (f.Control & 0x10) != 0
		f.NR = (f.Control >> 5) & 0x07
	} else {
		f.Type = FrameTypeU
	}

	return f, nil
}

// encodeAddress encodes an address (1, 2, or 4 bytes) with extension bits
func encodeAddress(addr []byte) []byte {
	if len(addr) != 1 && len(addr) != 2 && len(addr) != 4 {
		return nil
	}
	encoded := make([]byte, len(addr))
	for i := 0; i < len(addr); i++ {
		encoded[i] = addr[i] << 1
	}
	if len(addr) > 0 {
		encoded[len(addr)-1] |= 0x01
	}
	return encoded
}

// decodeAddress decodes an address (1, 2, or 4 bytes) with extension bits
func decodeAddress(data []byte) ([]byte, int) {
	var addr []byte
	for i := 0; i < len(data); i++ {
		addr = append(addr, data[i])
		if data[i]&0x01 == 1 {
			length := len(addr)
			if length != 1 && length != 2 && length != 4 {
				return nil, 0
			}
			decoded := make([]byte, length)
			for j := 0; j < length; j++ {
				decoded[j] = addr[j] >> 1
			}
			return decoded, length
		}
	}
	return nil, 0
}

// hasInformation checks if the frame has an information field
func hasInformation(control byte) bool {
	if control&0x01 == 0 {
		return true
	} else if control&0x03 == 0x01 {
		return false
	} else {
		switch control {
		case UFrameUI, UFrameFRMR:
			return true
		default:
			return false
		}
	}
}
