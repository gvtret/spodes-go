package axdr

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Decode decodes A-XDR data into a Go value as per IEC 62056-6-2 and СТО 34.01-5.1-006-2023.
// It supports primitive types, custom date/time types, bit strings, BCD, arrays, structures, and compact arrays.
// Returns the decoded value or an error if the data is invalid or unsupported.
func Decode(data []byte) (interface{}, error) {
	reader := bytes.NewReader(data)
	return decodeValue(reader)
}

// decodeFunc defines a function signature for type-specific decoding.
type decodeFunc func(reader *bytes.Reader) (interface{}, error)

// decodeDispatch maps tags to their decoding functions for performance optimization.
// This avoids switch statements for common types, improving decoding speed.
var decodeDispatch map[Tag]decodeFunc

// init initializes the decodeDispatch map to avoid initialization cycles.
func init() {
	decodeDispatch = map[Tag]decodeFunc{
		TagNull: func(reader *bytes.Reader) (interface{}, error) {
			return nil, nil
		},
		TagBoolean:                    decodeBoolean,
		TagInteger:                    decodeInt8,
		TagDeltaInteger:               decodeInt8,
		TagLong:                       decodeInt16,
		TagDeltaLong:                  decodeInt16,
		TagUnsigned:                   decodeUint8,
		TagDeltaUnsigned:              decodeUint8,
		TagLongUnsigned:               decodeUint16,
		TagDeltaLongUnsigned:          decodeUint16,
		TagDoubleLong:                 decodeInt32,
		TagDeltaDoubleLong:            decodeInt32,
		TagDoubleLongU:                decodeUint32,
		TagDeltaDoubleLongUnsigned:    decodeUint32,
		TagLong64:                     decodeInt64,
		TagLong64U:                    decodeUint64,
		TagFloat32:                    decodeFloat32,
		TagFloat64:                    decodeFloat64,
		TagOctetString:                decodeOctetString,
		TagVisibleString:              decodeVisibleString,
		TagBitString:                  decodeBitString,
		TagBCD:                        decodeBCD,
		TagDate:                       decodeDate,
		TagTime:                       decodeTime,
		TagDateTime:                   decodeDateTime,
		TagArray:                      decodeArray,
		TagStructure:                  decodeStructure,
		TagCompactArray:               decodeCompactArray,
	}
}

// decodeValue decodes a single A-XDR value based on its tag.
// It uses a dispatch table for performance, falling back to an error for unsupported tags.
// Returns the decoded value or an error if the tag is unsupported or data is invalid.
func decodeValue(reader *bytes.Reader) (interface{}, error) {
	if reader.Len() == 0 {
		return nil, fmt.Errorf("empty data")
	}
	tagByte, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read tag: %v", err)
	}
	tag := Tag(tagByte)

	// Use dispatch table for decoding.
	if decodeFn, ok := decodeDispatch[tag]; ok {
		return decodeFn(reader)
	}
	return nil, fmt.Errorf("unsupported tag: 0x%02x", tag)
}

// decodeBoolean decodes a boolean value (TagBoolean).
// Expects a single byte: 0 (false) or 1 (true).
// Returns the boolean value or an error if reading fails.
func decodeBoolean(reader *bytes.Reader) (interface{}, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to decode boolean: %v", err)
	}
	return b != 0, nil
}

// decodeInt8 decodes an 8-bit signed integer (TagInteger, TagDeltaInteger).
// Range: -128 to 127.
// Returns the integer or an error if reading fails.
func decodeInt8(reader *bytes.Reader) (interface{}, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to decode int8: %v", err)
	}
	return int8(b), nil
}

// decodeInt16 decodes a 16-bit signed integer (TagLong, TagDeltaLong).
// Range: -32,768 to 32,767.
// Returns the integer or an error if reading fails.
func decodeInt16(reader *bytes.Reader) (interface{}, error) {
	var val int16
	if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
		return nil, fmt.Errorf("failed to decode int16: %v", err)
	}
	return val, nil
}

// decodeUint8 decodes an 8-bit unsigned integer (TagUnsigned, TagDeltaUnsigned).
// Range: 0 to 255.
// Returns the integer or an error if reading fails.
func decodeUint8(reader *bytes.Reader) (interface{}, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to decode uint8: %v", err)
	}
	return uint8(b), nil
}

// decodeUint16 decodes a 16-bit unsigned integer (TagLongUnsigned, TagDeltaLongUnsigned).
// Range: 0 to 65,535.
// Returns the integer or an error if reading fails.
func decodeUint16(reader *bytes.Reader) (interface{}, error) {
	var val uint16
	if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
		return nil, fmt.Errorf("failed to decode uint16: %v", err)
	}
	return val, nil
}

// decodeInt32 decodes a 32-bit signed integer (TagDoubleLong, TagDeltaDoubleLong).
// Range: -2,147,483,648 to 2,147,483,647.
// Returns the integer or an error if reading fails.
func decodeInt32(reader *bytes.Reader) (interface{}, error) {
	var val int32
	if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
		return nil, fmt.Errorf("failed to decode int32: %v", err)
	}
	return val, nil
}

// decodeUint32 decodes a 32-bit unsigned integer (TagDoubleLongU, TagDeltaDoubleLongUnsigned).
// Range: 0 to 4,294,967,295.
// Returns the integer or an error if reading fails.
func decodeUint32(reader *bytes.Reader) (interface{}, error) {
	var val uint32
	if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
		return nil, fmt.Errorf("failed to decode uint32: %v", err)
	}
	return val, nil
}

// decodeInt64 decodes a 64-bit signed integer (TagLong64).
// Range: -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807.
// Returns the integer or an error if reading fails.
func decodeInt64(reader *bytes.Reader) (interface{}, error) {
	var val int64
	if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
		return nil, fmt.Errorf("failed to decode int64: %v", err)
	}
	return val, nil
}

// decodeUint64 decodes a 64-bit unsigned integer (TagLong64U).
// Range: 0 to 18,446,744,073,709,551,615.
// Returns the integer or an error if reading fails.
func decodeUint64(reader *bytes.Reader) (interface{}, error) {
	var val uint64
	if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
		return nil, fmt.Errorf("failed to decode uint64: %v", err)
	}
	return val, nil
}

// decodeFloat32 decodes a 32-bit floating-point number (TagFloat32, IEEE 754).
// Returns the float or an error if reading fails.
func decodeFloat32(reader *bytes.Reader) (interface{}, error) {
	var val float32
	if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
		return nil, fmt.Errorf("failed to decode float32: %v", err)
	}
	return val, nil
}

// decodeFloat64 decodes a 64-bit floating-point number (TagFloat64, IEEE 754).
// Returns the float or an error if reading fails.
func decodeFloat64(reader *bytes.Reader) (interface{}, error) {
	var val float64
	if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
		return nil, fmt.Errorf("failed to decode float64: %v", err)
	}
	return val, nil
}

// decodeOctetString decodes an octet-string (TagOctetString).
// Expects a length byte (0-255) followed by the byte sequence.
// Returns an empty byte slice for length 0, or the byte slice, or an error if reading fails.
func decodeOctetString(reader *bytes.Reader) (interface{}, error) {
	length, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet-string length: %v", err)
	}
	if length == 0 {
		return []byte{}, nil
	}
	data := make([]byte, length)
	if _, err := reader.Read(data); err != nil {
		return nil, fmt.Errorf("failed to decode octet-string: %v", err)
	}
	return data, nil
}

// decodeVisibleString decodes a visible-string (TagVisibleString).
// Expects a length byte (0-255) followed by ASCII bytes.
// Returns an empty string for length 0, or the string, or an error if reading fails.
func decodeVisibleString(reader *bytes.Reader) (interface{}, error) {
	length, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read visible-string length: %v", err)
	}
	if length == 0 {
		return "", nil
	}
	data := make([]byte, length)
	if _, err := reader.Read(data); err != nil {
		return nil, fmt.Errorf("failed to decode visible-string: %v", err)
	}
	return string(data), nil
}

// decodeBitString decodes a BitString value (TagBitString) per IEC 62056-6-2.
// Expects a length byte (0-255, number of bits) followed by ceiling(length/8) bytes.
// Unused bits in the last byte are expected to be zero.
// Returns a BitString struct or an error if the data is invalid.
func decodeBitString(reader *bytes.Reader) (interface{}, error) {
	length, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read bitstring length: %v", err)
	}
	expectedBytes := (length + 7) / 8
	data := make([]byte, expectedBytes)
	if _, err := reader.Read(data); err != nil {
		return nil, fmt.Errorf("failed to decode bitstring: %v", err)
	}
	bs := BitString{
		Bits:   data,
		Length: length,
	}
	if err := bs.Validate(); err != nil {
		return nil, fmt.Errorf("invalid bitstring: %v", err)
	}
	return bs, nil
}

// decodeBCD decodes a BCD value (TagBCD) per IEC 62056-6-2.
// Expects a length byte (0-255, number of digits) followed by ceiling(length/2) bytes,
// each encoding two decimal digits (high nibble first).
// Returns a BCD struct or an error if the data is invalid.
func decodeBCD(reader *bytes.Reader) (interface{}, error) {
	length, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read BCD length: %v", err)
	}
	expectedBytes := (length + 1) / 2
	data := make([]byte, expectedBytes)
	if _, err := reader.Read(data); err != nil {
		return nil, fmt.Errorf("failed to decode BCD: %v", err)
	}
	// Unpack bytes into digits.
	digits := make([]byte, length)
	for i := 0; i < int(length); i++ {
		byteIdx := i / 2
		if i%2 == 0 {
			digits[i] = (data[byteIdx] >> 4) & 0x0F
		} else {
			digits[i] = data[byteIdx] & 0x0F
		}
	}
	bcd := BCD{Digits: digits}
	if err := bcd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid BCD: %v", err)
	}
	return bcd, nil
}

// decodeDate decodes a Date value (TagDate, 5 bytes) per IEC 62056-6-2 clause 4.1.6.1.
// Expects 5 bytes: year (2), month (1), day (1), day of week (1).
// Returns a Date struct or an error if the data is invalid.
func decodeDate(reader *bytes.Reader) (interface{}, error) {
	data := make([]byte, 5)
	if _, err := reader.Read(data); err != nil {
		return nil, fmt.Errorf("failed to decode date: %v", err)
	}
	d := Date{
		Year:      binary.BigEndian.Uint16(data[0:2]),
		Month:     data[2],
		Day:       data[3],
		DayOfWeek: data[4],
	}
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("invalid date: %v", err)
	}
	return d, nil
}

// decodeTime decodes a Time value (TagTime, 4 bytes) per IEC 62056-6-2 clause 4.1.6.1.
// Expects 4 bytes: hour (1), minute (1), second (1), hundredths (1).
// Returns a Time struct or an error if the data is invalid.
func decodeTime(reader *bytes.Reader) (interface{}, error) {
	data := make([]byte, 4)
	if _, err := reader.Read(data); err != nil {
		return nil, fmt.Errorf("failed to decode time: %v", err)
	}
	t := Time{
		Hour:       data[0],
		Minute:     data[1],
		Second:     data[2],
		Hundredths: data[3],
	}
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("invalid time: %v", err)
	}
	return t, nil
}

// decodeDateTime decodes a DateTime value (TagDateTime, 12 bytes) per IEC 62056-6-2 clause 4.1.6.1.
// Expects 12 bytes: year (2), month (1), day (1), day of week (1), hour (1), minute (1),
// second (1), hundredths (1), deviation (2), clock status (1).
// Returns a DateTime struct or an error if the data is invalid.
func decodeDateTime(reader *bytes.Reader) (interface{}, error) {
	data := make([]byte, 12)
	if _, err := reader.Read(data); err != nil {
		return nil, fmt.Errorf("failed to decode datetime: %v", err)
	}
	dt := DateTime{
		Date: Date{
			Year:      binary.BigEndian.Uint16(data[0:2]),
			Month:     data[2],
			Day:       data[3],
			DayOfWeek: data[4],
		},
		Time: Time{
			Hour:       data[5],
			Minute:     data[6],
			Second:     data[7],
			Hundredths: data[8],
		},
		Deviation:   int16(binary.BigEndian.Uint16(data[9:11])),
		ClockStatus: data[11],
	}
	if err := dt.Validate(); err != nil {
		return nil, fmt.Errorf("invalid datetime: %v", err)
	}
	return dt, nil
}

// decodeArray decodes an A-XDR Array (TagArray).
// Expects a length byte (0-255) followed by encoded elements, each with its own tag.
// Returns a slice of interfaces or an error if decoding fails.
func decodeArray(reader *bytes.Reader) (interface{}, error) {
	length, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read array length: %v", err)
	}
	result := make([]interface{}, length)
	// Use range loop for idiomatic Go iteration.
	for i := range result {
		val, err := decodeValue(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode array element %d: %v", i, err)
		}
		result[i] = val
	}
	return result, nil
}

// decodeStructure decodes an A-XDR Structure (TagStructure).
// Expects a length byte (0-255) followed by encoded fields, each with its own tag.
// Returns a slice of interfaces or an error if decoding fails.
func decodeStructure(reader *bytes.Reader) (interface{}, error) {
	length, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read structure length: %v", err)
	}
	result := make([]interface{}, length)
	// Use range loop for idiomatic Go iteration.
	for i := range result {
		val, err := decodeValue(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode structure field %d: %v", i, err)
		}
		result[i] = val
	}
	return result, nil
}
