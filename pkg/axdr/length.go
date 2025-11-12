package axdr

import (
	"bytes"
	"fmt"
)

// writeAXDRLength encodes a non-negative length using the AXDR variable-length encoding.
// Lengths up to 127 are encoded in a single byte. Longer lengths use the high bit as an
// extension indicator and store the big-endian length across the following bytes.
func writeAXDRLength(buf *bytes.Buffer, length int) error {
	if length < 0 {
		return fmt.Errorf("negative length: %d", length)
	}
	if length <= 0x7F {
		buf.WriteByte(byte(length))
		return nil
	}

	value := uint64(length)
	var tmp [8]byte
	i := len(tmp)
	for value > 0 {
		i--
		tmp[i] = byte(value & 0xFF)
		value >>= 8
	}
	lengthBytes := tmp[i:]
	if len(lengthBytes) == 0 {
		lengthBytes = []byte{0}
	}
	if len(lengthBytes) > 0x7F {
		return fmt.Errorf("length %d exceeds AXDR encoding limits", length)
	}

	buf.WriteByte(0x80 | byte(len(lengthBytes)))
	buf.Write(lengthBytes)
	return nil
}

// readAXDRLength decodes an AXDR variable-length length field from the provided reader.
// It mirrors writeAXDRLength, supporting multi-byte definite-length encodings.
func readAXDRLength(reader *bytes.Reader) (int, error) {
	first, err := reader.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("failed to read length: %w", err)
	}
	if first&0x80 == 0 {
		return int(first), nil
	}

	numBytes := int(first & 0x7F)
	if numBytes == 0 {
		return 0, fmt.Errorf("indefinite lengths are not supported")
	}
	if numBytes > 8 {
		return 0, fmt.Errorf("length uses %d bytes, exceeds supported maximum of 8", numBytes)
	}

	var length uint64
	for i := 0; i < numBytes; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("failed to read length byte %d: %w", i, err)
		}
		length = (length << 8) | uint64(b)
	}

	maxInt := uint64(int(^uint(0) >> 1))
	if length > maxInt {
		return 0, fmt.Errorf("length %d exceeds supported maximum of %d", length, maxInt)
	}

	return int(length), nil
}
