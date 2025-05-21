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
	if len(data) == 0 {
		return []byte{}
	}
	// Estimate max size: original + 1 stuffed bit per 5 bits + 1 for partial byte.
	// len(data) * 8 bits / 5 bits_group = len(data) * 1.6. Add original len(data).
	// So, roughly len(data) * 2.6 for bits, or len(data) * 1.2 for bytes.
	// Max stuffed size is len(data) + len(data)/5. Add 1 for the last byte.
	estimatedMaxSize := len(data) + (len(data) / 5) + 1
	result := make([]byte, 0, estimatedMaxSize)

	var currentWriteByte byte
	bitWritePos := 0 // Counts 0 to 7 for currentWriteByte, MSB-first fill
	ones := 0

	for _, b := range data { // Iterate over each input byte
		for i := 7; i >= 0; i-- { // Iterate over each bit in the input byte (MSB to LSB)
			bit := (b >> uint(i)) & 1

			// Add the current bit to currentWriteByte
			currentWriteByte = (currentWriteByte << 1) | bit
			bitWritePos++

			if bit == 1 {
				ones++
			} else {
				ones = 0
			}

			if bitWritePos == 8 {
				result = append(result, currentWriteByte)
				currentWriteByte = 0
				bitWritePos = 0
			}

			if ones == 5 {
				// Stuff a 0 bit
				currentWriteByte <<= 1 // Shift left to make space for the 0, it's already 0
				bitWritePos++
				ones = 0
				if bitWritePos == 8 {
					result = append(result, currentWriteByte)
					currentWriteByte = 0
					bitWritePos = 0
				}
			}
		}
	}

	// If there are remaining bits in currentWriteByte, pad and append
	if bitWritePos > 0 {
		currentWriteByte <<= (8 - uint(bitWritePos)) // Pad with trailing zeros
		result = append(result, currentWriteByte)
	}

	return result
}

// bitUnstuff removes bit stuffing from data
func bitUnstuff(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}
	// Output can't be larger than input.
	result := make([]byte, 0, len(data))

	var outputCurrentByte byte
	outputBitPos := 0 // Counts 0 to 7 for outputCurrentByte, MSB-first fill
	ones := 0         // Consecutive ones count from input stream

	for _, b := range data { // Iterate over each input byte (stuffed data)
		for i := 7; i >= 0; i-- { // Iterate over each bit (MSB to LSB)
			bit := (b >> uint(i)) & 1

			if ones == 5 {
				if bit == 0 { // This is a stuffed '0' bit
					ones = 0   // Reset counter
					continue // Skip this bit (do not add to output)
				}
				// If bit is 1 after five 1s, it's an error (flag sequence in data that wasn't a flag)
				return nil, errors.New("invalid bit stuffing")
			}

			// Add the bit to outputCurrentByte
			outputCurrentByte = (outputCurrentByte << 1) | bit
			outputBitPos++

			if bit == 1 {
				ones++
			} else {
				ones = 0
			}

			if outputBitPos == 8 {
				result = append(result, outputCurrentByte)
				outputCurrentByte = 0
				outputBitPos = 0
			}
		}
	}

	// After processing all input bytes, if outputBitPos > 0, these are
	// valid data bits that form an incomplete byte. This implies the original
	// unstuffed data was not a multiple of 8 bits long.
	// Standard HDLC typically has byte-aligned information fields.
	// The previous version correctly did *not* append this potentially partial byte.
	// This is correct because if the original data was, say, 9 bits, it would be
	// transmitted as 2 bytes (with 7 padding bits). Unstuffing should yield only
	// the meaningful 9 bits, and if byte-aligned output is expected, this would
	// mean 1 full byte and 1 bit for the next, which isn't how this usually works.
	// The length of useful data is typically known from higher layers or frame length field.
	// For this raw function, returning only full bytes is safer and matches prior fix.
	return result, nil
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
	if len(unstuffed) < length+4 {
		return nil, errors.New("frame too short")
	}

	// Trim to exact length to handle padding from bit stuffing
	unstuffed = unstuffed[0 : length+4]

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