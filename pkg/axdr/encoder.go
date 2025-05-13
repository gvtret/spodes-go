package axdr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// Encode encodes a value into A-XDR format as per IEC 62056-6-2 and СТО 34.01-5.1-006-2023.
// It supports primitive types, custom date/time types, bit strings, BCD, arrays, structures, and compact arrays.
// Returns the encoded byte slice or an error if the value cannot be encoded.
func Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeValue(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encodeFunc defines a function signature for type-specific encoding.
type encodeFunc func(buf *bytes.Buffer, v interface{}) error

// encodeDispatch maps Go types to their encoding functions for performance optimization.
// This avoids reflection for common types, improving encoding speed.
var encodeDispatch map[reflect.Type]encodeFunc

// init initializes the encodeDispatch map to avoid initialization cycles.
func init() {
	encodeDispatch = map[reflect.Type]encodeFunc{
		reflect.TypeOf(false): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagBoolean, func() { buf.WriteByte(boolToByte(v.(bool))) })
		},
		reflect.TypeOf(int8(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagDeltaInteger, func() { buf.WriteByte(byte(v.(int8))) })
		},
		reflect.TypeOf(int16(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagDeltaLong, func() { binary.Write(buf, binary.BigEndian, v.(int16)) })
		},
		reflect.TypeOf(uint8(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagDeltaUnsigned, func() { buf.WriteByte(v.(uint8)) })
		},
		reflect.TypeOf(uint16(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagDeltaLongUnsigned, func() { binary.Write(buf, binary.BigEndian, v.(uint16)) })
		},
		reflect.TypeOf(int32(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagDeltaDoubleLong, func() { binary.Write(buf, binary.BigEndian, v.(int32)) })
		},
		reflect.TypeOf(uint32(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagDeltaDoubleLongUnsigned, func() { binary.Write(buf, binary.BigEndian, v.(uint32)) })
		},
		reflect.TypeOf(int64(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagLong64, func() { binary.Write(buf, binary.BigEndian, v.(int64)) })
		},
		reflect.TypeOf(uint64(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagLong64U, func() { binary.Write(buf, binary.BigEndian, v.(uint64)) })
		},
		reflect.TypeOf(float32(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagFloat32, func() { binary.Write(buf, binary.BigEndian, v.(float32)) })
		},
		reflect.TypeOf(float64(0)): func(buf *bytes.Buffer, v interface{}) error {
			return encodePrimitive(buf, reflect.ValueOf(v), TagFloat64, func() { binary.Write(buf, binary.BigEndian, v.(float64)) })
		},
		reflect.TypeOf(""): func(buf *bytes.Buffer, v interface{}) error {
			return encodeString(buf, reflect.ValueOf(v), TagVisibleString)
		},
		reflect.TypeOf([]byte{}): func(buf *bytes.Buffer, v interface{}) error {
			return encodeOctetString(buf, reflect.ValueOf(v))
		},
		reflect.TypeOf(Date{}): func(buf *bytes.Buffer, v interface{}) error {
			return encodeDate(buf, v.(Date))
		},
		reflect.TypeOf(Time{}): func(buf *bytes.Buffer, v interface{}) error {
			return encodeTime(buf, v.(Time))
		},
		reflect.TypeOf(DateTime{}): func(buf *bytes.Buffer, v interface{}) error {
			return encodeDateTime(buf, v.(DateTime))
		},
		reflect.TypeOf(BitString{}): func(buf *bytes.Buffer, v interface{}) error {
			return encodeBitString(buf, v.(BitString))
		},
		reflect.TypeOf(BCD{}): func(buf *bytes.Buffer, v interface{}) error {
			return encodeBCD(buf, v.(BCD))
		},
		reflect.TypeOf(CompactArray{}): func(buf *bytes.Buffer, v interface{}) error {
			return encodeCompactArray(buf, v.(CompactArray))
		},
	}
}

// encodeValue handles the top-level encoding logic, using a dispatch table for performance.
// Falls back to reflection for unsupported types (arrays, structs).
// Returns an error if the value is unsupported or invalid.
func encodeValue(buf *bytes.Buffer, v interface{}) error {
	if v == nil {
		// Encode null as a single byte.
		buf.WriteByte(byte(TagNull))
		return nil
	}

	// Check dispatch table for type-specific encoding.
	if encodeFn, ok := encodeDispatch[reflect.TypeOf(v)]; ok {
		return encodeFn(buf, v)
	}

	// Fallback to reflection for arrays and structs.
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		if rv.Type().Elem().Kind() != reflect.Uint8 {
			return encodeArray(buf, rv)
		}
		return encodeOctetString(buf, rv)
	case reflect.Array:
		return encodeArray(buf, rv)
	case reflect.Struct:
		return encodeStructure(buf, rv)
	default:
		return fmt.Errorf("unsupported type: %v", reflect.TypeOf(v))
	}
}

// encodePrimitive encodes a primitive type with the specified tag and write function.
// The write function handles the binary representation of the value.
// Returns an error if writing fails.
func encodePrimitive(buf *bytes.Buffer, v reflect.Value, tag Tag, writeFunc func()) error {
	// Write the tag followed by the value.
	buf.WriteByte(byte(tag))
	writeFunc()
	return nil
}

// boolToByte converts a boolean to a byte: 1 for true, 0 for false.
func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// encodeString encodes a string type (e.g., VisibleString) with the specified tag.
// The string is encoded as a length-prefixed ASCII byte sequence.
// Returns an error if the string length exceeds 255 bytes.
func encodeString(buf *bytes.Buffer, v reflect.Value, tag Tag) error {
	data := []byte(v.String())
	if len(data) > 255 {
		return fmt.Errorf("string length %d exceeds maximum of 255", len(data))
	}
	// Write tag, length, and string data.
	buf.WriteByte(byte(tag))
	buf.WriteByte(byte(len(data)))
	buf.Write(data)
	return nil
}

// encodeOctetString encodes an octet-string (byte slice) with TagOctetString.
// The byte slice is encoded as a length-prefixed sequence.
// Returns an error if the length exceeds 255 bytes.
func encodeOctetString(buf *bytes.Buffer, v reflect.Value) error {
	data := v.Bytes()
	if len(data) > 255 {
		return fmt.Errorf("octet-string length %d exceeds maximum of 255", len(data))
	}
	// Write tag, length, and byte data.
	buf.WriteByte(byte(TagOctetString))
	buf.WriteByte(byte(len(data)))
	buf.Write(data)
	return nil
}

// encodeBitString encodes a BitString value with TagBitString per IEC 62056-6-2.
// The format is: tag (0x04), length (1 byte, number of bits), followed by ceiling(length/8) bytes.
// Unused bits in the last byte are padded with zeros.
// Returns an error if the BitString is invalid.
func encodeBitString(buf *bytes.Buffer, bs BitString) error {
	if err := bs.Validate(); err != nil {
		return fmt.Errorf("invalid bitstring: %v", err)
	}
	// Write tag and bit length.
	buf.WriteByte(byte(TagBitString))
	buf.WriteByte(bs.Length)
	// Write bit data.
	buf.Write(bs.Bits)
	return nil
}

// encodeBCD encodes a BCD value with TagBCD per IEC 62056-6-2.
// The format is: tag (0x0D), length (1 byte, number of digits), followed by ceiling(length/2) bytes,
// each encoding two decimal digits (high nibble first).
// Returns an error if the BCD is invalid.
func encodeBCD(buf *bytes.Buffer, bcd BCD) error {
	if err := bcd.Validate(); err != nil {
		return fmt.Errorf("invalid BCD: %v", err)
	}
	// Write tag and digit count.
	length := len(bcd.Digits)
	buf.WriteByte(byte(TagBCD))
	buf.WriteByte(byte(length))
	// Pack digits into bytes (two per byte, high nibble first).
	for i := 0; i < length; i += 2 {
		var b byte
		b = bcd.Digits[i] << 4
		if i+1 < length {
			b |= bcd.Digits[i+1]
		}
		buf.WriteByte(b)
	}
	return nil
}

// encodeDate encodes a Date value as a 5-byte sequence per IEC 62056-6-2 clause 4.1.6.1.
// The format is: year (2 bytes), month (1 byte), day (1 byte), day of week (1 byte).
// Returns an error if the Date is invalid.
func encodeDate(buf *bytes.Buffer, d Date) error {
	if err := d.Validate(); err != nil {
		return fmt.Errorf("invalid date: %v", err)
	}
	// Write tag and date fields.
	buf.WriteByte(byte(TagDate))
	data := []byte{
		byte(d.Year >> 8), byte(d.Year & 0xFF),
		d.Month, d.Day, d.DayOfWeek,
	}
	buf.Write(data)
	return nil
}

// encodeTime encodes a Time value as a 4-byte sequence per IEC 62056-6-2 clause 4.1.6.1.
// The format is: hour (1 byte), minute (1 byte), second (1 byte), hundredths (1 byte).
// Returns an error if the Time is invalid.
func encodeTime(buf *bytes.Buffer, t Time) error {
	if err := t.Validate(); err != nil {
		return fmt.Errorf("invalid time: %v", err)
	}
	// Write tag and time fields.
	buf.WriteByte(byte(TagTime))
	data := []byte{
		t.Hour, t.Minute, t.Second, t.Hundredths,
	}
	buf.Write(data)
	return nil
}

// encodeDateTime encodes a DateTime value as a 12-byte sequence per IEC 62056-6-2 clause 4.1.6.1.
// The format includes date (5 bytes), time (4 bytes), deviation (2 bytes), and clock status (1 byte).
// Returns an error if the DateTime is invalid.
func encodeDateTime(buf *bytes.Buffer, dt DateTime) error {
	if err := dt.Validate(); err != nil {
		return fmt.Errorf("invalid datetime: %v", err)
	}
	// Write tag and datetime fields.
	buf.WriteByte(byte(TagDateTime))
	data := []byte{
		byte(dt.Date.Year >> 8), byte(dt.Date.Year & 0xFF),
		dt.Date.Month, dt.Date.Day, dt.Date.DayOfWeek,
		dt.Time.Hour, dt.Time.Minute, dt.Time.Second, dt.Time.Hundredths,
		byte(dt.Deviation >> 8), byte(dt.Deviation),
		dt.ClockStatus,
	}
	buf.Write(data)
	return nil
}

// encodeArray encodes a slice or array as an A-XDR Array with TagArray.
// Each element is encoded with its own type tag.
// Returns an error if the array length exceeds 255 or if any element fails to encode.
func encodeArray(buf *bytes.Buffer, v reflect.Value) error {
	// Write tag and length.
	buf.WriteByte(byte(TagArray))
	length := v.Len()
	if length > 255 {
		return fmt.Errorf("array length %d exceeds maximum of 255", length)
	}
	buf.WriteByte(byte(length))
	// Encode each element using range loop for idiomatic Go.
	for i, val := range v.Slice(0, length).Interface().([]interface{}) {
		data, err := Encode(val)
		if err != nil {
			return fmt.Errorf("failed to encode array element %d: %v", i, err)
		}
		buf.Write(data)
	}
	return nil
}

// encodeStructure encodes a struct as an A-XDR Structure with TagStructure.
// Each field is encoded with its own type tag.
// Returns an error if the field count exceeds 255 or if any field fails to encode.
func encodeStructure(buf *bytes.Buffer, v reflect.Value) error {
	// Write tag and field count.
	buf.WriteByte(byte(TagStructure))
	length := v.NumField()
	if length > 255 {
		return fmt.Errorf("structure field count %d exceeds maximum of 255", length)
	}
	buf.WriteByte(byte(length))
	// Encode each field using range loop for idiomatic Go.
	for i := 0; i < length; i++ {
		data, err := Encode(v.Field(i).Interface())
		if err != nil {
			return fmt.Errorf("failed to encode structure field %d: %v", i, err)
		}
		buf.Write(data)
	}
	return nil
}
