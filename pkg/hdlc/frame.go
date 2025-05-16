package hdlc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

// Constants for HDLC protocol
const (
	FlagByte             = 0x7E        // Frame delimiter
	MaxWindowSize        = 7           // Maximum window size for sliding window
	MaxFrameSize         = 2048        // Maximum frame size in bytes
	InterOctetTimeout    = 200         // Inter-octet timeout in milliseconds
	InactivityTimeout    = 30 * 1000   // Inactivity timeout in milliseconds
	BroadcastAddress     = 0xFF        // Broadcast address for UI frames
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
	Format      uint16    // Frame format field (2 bytes)
	DA          []byte    // Destination Address (logical or physical server address)
	SA          []byte    // Source Address (client address)
	Control     byte      // Control field
	HCS         uint16    // Header Check Sequence
	Information []byte    // Information field
	FCS         uint16    // Frame Check Sequence
	Type        int       // Frame type (I, S, U)
	NS          uint8     // Send sequence number
	NR          uint8     // Receive sequence number
	PF          bool      // Poll/Final bit
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

// bitStuff applies bit stuffing to prevent flag sequence in data
func bitStuff(data []byte) []byte {
	var result []byte
	var currentByte byte
	bitPos := 0
	countOnes := 0
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bit := (b >> i) & 1
			currentByte = (currentByte << 1) | bit
			bitPos++
			if bit == 1 {
				countOnes++
				if countOnes == 5 {
					currentByte <<= 1 // Insert zero bit
					bitPos++
					countOnes = 0
				}
			} else {
				countOnes = 0
			}
			if bitPos == 8 {
				result = append(result, currentByte)
				currentByte = 0
				bitPos = 0
			}
		}
	}
	if bitPos > 0 {
		currentByte <<= (8 - bitPos)
		result = append(result, currentByte)
	}
	return result
}

// bitUnstuff removes bit stuffing from data
func bitUnstuff(data []byte) ([]byte, error) {
	var bits []byte
	countOnes := 0

	// Extract bits, removing stuffed zeros
	for _, b := range data {
		if err := processByteBits(b, &bits, &countOnes); err != nil {
			return nil, err
		}
	}

	// Convert bits to bytes
	return bitsToBytes(bits), nil
}

// processByteBits processes a single byte's bits for bit unstuffing
func processByteBits(b byte, bits *[]byte, countOnes *int) error {
	for i := 7; i >= 0; i-- {
		bit := (b >> i) & 1
		if *countOnes == 5 {
			if bit == 0 {
				*countOnes = 0
				continue
			}
			return errors.New("invalid bit stuffing")
		}
		*bits = append(*bits, byte(bit))
		if bit == 1 {
			*countOnes++
		} else {
			*countOnes = 0
		}
	}
	return nil
}

// bitsToBytes converts a slice of bits to bytes
func bitsToBytes(bits []byte) []byte {
	var result []byte
	for i := 0; i < len(bits); i += 8 {
		var byteVal byte
		for j := 0; j < 8 && i+j < len(bits); j++ {
			if bits[i+j] == 1 {
				byteVal |= 1 << (7 - j)
			}
		}
		result = append(result, byteVal)
	}
	return result
}

// EncodeFrame encodes an HDLC frame with bit stuffing and CRC
func EncodeFrame(da, sa []byte, control byte, info []byte, segmented bool) ([]byte, error) {
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
	length := len(encodedDA) + len(encodedSA) + 1
	if hasInfo {
		length += 2 + len(info)
	}
	if length > 2047 {
		return nil, errors.New("frame too long")
	}
	s := 0
	if segmented {
		s = 1
	}
	format := uint16(0xA<<12) | uint16(s<<11) | uint16(length&0x7FF)
	var header bytes.Buffer
	binary.Write(&header, binary.BigEndian, format)
	header.Write(encodedDA)
	header.Write(encodedSA)
	header.WriteByte(control)
	if hasInfo {
		hcs := calculateCRC16(header.Bytes())
		binary.Write(&header, binary.BigEndian, hcs)
	}
	var payload bytes.Buffer
	payload.Write(header.Bytes())
	if hasInfo {
		payload.Write(info)
	}
	fcs := calculateCRC16(payload.Bytes())
	binary.Write(&payload, binary.BigEndian, fcs)
	stuffed := bitStuff(payload.Bytes())
	var frame bytes.Buffer
	frame.WriteByte(FlagByte)
	frame.Write(stuffed)
	frame.WriteByte(FlagByte)
	return frame.Bytes(), nil
}

// DecodeFrame decodes an HDLC frame, validating CRC and structure
func DecodeFrame(frame []byte, interOctetTimeout time.Duration) (*HDLCFrame, error) {
	if len(frame) < 1 || frame[0] != FlagByte {
		return nil, errors.New("missing start flag")
	}
	if len(frame) > 1 && frame[len(frame)-1] != FlagByte {
		return nil, errors.New("incomplete frame received")
	}

	data := frame[1 : len(frame)-1]
	unstuffed, err := bitUnstuff(data)
	if err != nil {
		return nil, err
	}

	f, err := validateFrameStructure(unstuffed)
	if err != nil {
		return nil, err
	}

	return parseFrameControl(f, unstuffed)
}

// validateFrameStructure validates the frame's structure and CRC
func validateFrameStructure(unstuffed []byte) (*HDLCFrame, error) {
	if len(unstuffed) < 4 {
		return nil, errors.New("frame too short")
	}

	format := binary.BigEndian.Uint16(unstuffed[0:2])
	if (format>>12)&0xF != 0xA {
		return nil, errors.New("invalid format type")
	}

	length := int(format & 0x7FF)
	if len(unstuffed) != length+4 {
		return nil, errors.New("length mismatch")
	}

	fcsReceived := binary.BigEndian.Uint16(unstuffed[2+length : 2+length+2])
	if calculateCRC16(unstuffed[0:2+length]) != fcsReceived {
		return nil, errors.New("FCS mismatch")
	}

	dataPart := unstuffed[2 : 2+length]
	da, daLen := decodeAddress(dataPart)
	if daLen == 0 {
		return nil, errors.New("invalid destination address")
	}

	saStart := daLen
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

// parseFrameControl parses the control field and information data
func parseFrameControl(f *HDLCFrame, unstuffed []byte) (*HDLCFrame, error) {
	dataPart := unstuffed[2 : 2+int(f.Format&0x7FF)]
	controlStart := len(f.DA) + len(f.SA)
	hasInfo := hasInformation(f.Control)

	if hasInfo {
		hcsStart := controlStart + 1
		if hcsStart+2 > len(dataPart) {
			return nil, errors.New("missing HCS")
		}
		f.HCS = binary.BigEndian.Uint16(dataPart[hcsStart : hcsStart+2])
		header := unstuffed[0:2]
		header = append(header, dataPart[:controlStart+1]...)
		if calculateCRC16(header) != f.HCS {
			return nil, errors.New("HCS mismatch")
		}
		f.Information = dataPart[hcsStart+2:]
	} else if controlStart+1 < len(dataPart) {
		return nil, errors.New("unexpected data after control")
	}

	if f.Control&0x01 == 0 {
		f.Type = FrameTypeI
		f.NS = (f.Control >> 1) & 0x07
		f.PF = (f.Control & 0x10) != 0
		f.NR = (f.Control >> 5) & 0x07
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
		return nil // Invalid address length
	}
	encoded := make([]byte, len(addr))
	for i := 0; i < len(addr)-1; i++ {
		encoded[i] = addr[i] & 0xFE // Clear LSB for extension
	}
	if len(addr) > 0 {
		encoded[len(addr)-1] = addr[len(addr)-1] | 0x01 // Set LSB for last byte
	}
	return encoded
}

// decodeAddress decodes an address (1, 2, or 4 bytes) with extension bits
func decodeAddress(data []byte) ([]byte, int) {
	var addr []byte
	for i := 0; i < len(data); i++ {
		addr = append(addr, data[i])
		if data[i]&0x01 == 1 {
			if len(addr) != 1 && len(addr) != 2 && len(addr) != 4 {
				return nil, 0 // Invalid address length
			}
			return addr, i + 1
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
