package axdr

import (
	"bytes"
	"fmt"
)

// CompactArray represents an A-XDR CompactArray (TagCompactArray) as defined in IEC 62056-6-2.
// It contains a single type tag for all elements, optimizing space by omitting individual tags.
// The TypeTag specifies the type of all Values, which must be compatible with the tag.
type CompactArray struct {
	TypeTag Tag           // Type tag for all elements (e.g., TagInteger, TagDate).
	Values  []interface{} // Array of values, all of the same type as specified by TypeTag.
}

// EncodeCompactArray encodes a CompactArray into A-XDR format with TagCompactArray.
// The encoding format is: tag (0x13), length (1 byte, 0–255), type tag (1 byte), followed by
// element data without individual tags. Returns the encoded bytes or an error if encoding fails.
func encodeCompactArray(buf *bytes.Buffer, ca CompactArray) error {
	// Write compact array tag.
	buf.WriteByte(byte(TagCompactArray))
	// Write length (number of elements).
	length := len(ca.Values)
	if length > 255 {
		return fmt.Errorf("compact array length %d exceeds maximum of 255", length)
	}
	buf.WriteByte(byte(length))
	// Write the single type tag for all elements.
	buf.WriteByte(byte(ca.TypeTag))
	// Encode each element using range loop for idiomatic Go.
	for i, v := range ca.Values {
		data, err := Encode(v)
		if err != nil {
			return fmt.Errorf("failed to encode compact array element %d: %v", i, err)
		}
		if len(data) > 0 {
			// Skip the tag in the encoded data, as TypeTag is already written.
			buf.Write(data[1:])
		}
	}
	return nil
}

// decodeCompactArray decodes an A-XDR CompactArray (TagCompactArray).
// Expects a length byte (0–255), a type tag, and element data without individual tags.
// Returns a CompactArray struct or an error if decoding fails or the type tag is unsupported.
func decodeCompactArray(reader *bytes.Reader) (interface{}, error) {
	// Read length.
	length, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read compact array length: %v", err)
	}
	// Read type tag for all elements.
	typeTag, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read compact array type tag: %v", err)
	}
	// Decode each element based on the type tag.
	values := make([]interface{}, length)
	for i := 0; i < int(length); i++ {
		var val interface{}
		switch Tag(typeTag) {
		case TagBoolean:
			val, err = decodeBoolean(reader)
		case TagInteger, TagDeltaInteger:
			val, err = decodeInt8(reader)
		case TagLong, TagDeltaLong:
			val, err = decodeInt16(reader)
		case TagUnsigned, TagDeltaUnsigned:
			val, err = decodeUint8(reader)
		case TagLongUnsigned, TagDeltaLongUnsigned:
			val, err = decodeUint16(reader)
		case TagDoubleLong, TagDeltaDoubleLong:
			val, err = decodeInt32(reader)
		case TagDoubleLongU, TagDeltaDoubleLongUnsigned:
			val, err = decodeUint32(reader)
		case TagLong64:
			val, err = decodeInt64(reader)
		case TagLong64U:
			val, err = decodeUint64(reader)
		case TagFloat32:
			val, err = decodeFloat32(reader)
		case TagFloat64:
			val, err = decodeFloat64(reader)
		case TagOctetString:
			val, err = decodeOctetString(reader)
		case TagVisibleString:
			val, err = decodeVisibleString(reader)
		case TagBitString:
			val, err = decodeBitString(reader)
		case TagBCD:
			val, err = decodeBCD(reader)
		case TagDate:
			val, err = decodeDate(reader)
		case TagTime:
			val, err = decodeTime(reader)
		case TagDateTime:
			val, err = decodeDateTime(reader)
		default:
			return nil, fmt.Errorf("unsupported compact array type tag: 0x%02x", typeTag)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode compact array element %d: %v", i, err)
		}
		values[i] = val
	}
	return CompactArray{TypeTag: Tag(typeTag), Values: values}, nil
}
