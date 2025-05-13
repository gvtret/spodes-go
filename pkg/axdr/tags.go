package axdr

// Tag represents an A-XDR type tag as defined in IEC 62056-6-2 (Table 3) and
// СТО 34.01-5.1-006-2023 (Table 7.2). Each tag identifies the data type and its
// encoding format in the A-XDR binary representation.
type Tag byte

// A-XDR tag constants with their corresponding data types and value ranges.
// Ranges are specified as per the standards, with special values noted where applicable.
const (
	// TagNull represents a null value (no data).
	// Encoding: Single byte (0x00).
	TagNull Tag = 0

	// TagArray represents an array of elements, each with its own type tag.
	// Encoding: Tag (0x01), length (1 byte, 0–255), followed by encoded elements.
	TagArray Tag = 1

	// TagStructure represents a structure of heterogeneous elements, each with its own type tag.
	// Encoding: Tag (0x02), length (1 byte, 0–255), followed by encoded elements.
	TagStructure Tag = 2

	// TagBoolean represents a boolean value.
	// Range: 0 (false), 1 (true).
	// Encoding: Tag (0x03), value (1 byte).
	TagBoolean Tag = 3

	// TagBitString represents a bit string (not implemented due to unspecified format).
	// Encoding: Tag (0x04), length (in bits), followed by byte sequence.
	TagBitString Tag = 4

	// TagDoubleLong represents a 32-bit signed integer.
	// Range: -2,147,483,648 to 2,147,483,647.
	// Encoding: Tag (0x05), value (4 bytes, big-endian).
	TagDoubleLong Tag = 5

	// TagDoubleLongU represents a 32-bit unsigned integer.
	// Range: 0 to 4,294,967,295.
	// Encoding: Tag (0x06), value (4 bytes, big-endian).
	TagDoubleLongU Tag = 6

	// TagOctetString represents a sequence of bytes.
	// Range: 0–255 bytes.
	// Encoding: Tag (0x09), length (1 byte), followed by bytes.
	TagOctetString Tag = 9

	// TagVisibleString represents an ASCII string.
	// Range: 0–255 characters (ASCII).
	// Encoding: Tag (0x0A), length (1 byte), followed by ASCII bytes.
	TagVisibleString Tag = 10

	// TagUTF8String represents a UTF-8 encoded string (not implemented due to unspecified format).
	// Encoding: Tag (0x0C), length (1 byte), followed by UTF-8 bytes.
	TagUTF8String Tag = 12

	// TagBCD represents a binary-coded decimal (not implemented due to unspecified format).
	// Encoding: Tag (0x0D), followed by BCD digits.
	TagBCD Tag = 13

	// TagInteger represents an 8-bit signed integer.
	// Range: -128 to 127.
	// Encoding: Tag (0x0F), value (1 byte).
	TagInteger Tag = 15

	// TagLong represents a 16-bit signed integer.
	// Range: -32,768 to 32,767.
	// Encoding: Tag (0x10), value (2 bytes, big-endian).
	TagLong Tag = 16

	// TagUnsigned represents an 8-bit unsigned integer.
	// Range: 0 to 255.
	// Encoding: Tag (0x11), value (1 byte).
	TagUnsigned Tag = 17

	// TagLongUnsigned represents a 16-bit unsigned integer.
	// Range: 0 to 65,535.
	// Encoding: Tag (0x12), value (2 bytes, big-endian).
	TagLongUnsigned Tag = 18

	// TagCompactArray represents a compact array with a single type tag for all elements.
	// Encoding: Tag (0x13), length (1 byte, 0–255), type tag (1 byte), followed by elements without individual tags.
	TagCompactArray Tag = 19

	// TagLong64 represents a 64-bit signed integer.
	// Range: -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807.
	// Encoding: Tag (0x14), value (8 bytes, big-endian).
	TagLong64 Tag = 20

	// TagLong64U represents a 64-bit unsigned integer.
	// Range: 0 to 18,446,744,073,709,551,615.
	// Encoding: Tag (0x15), value (8 bytes, big-endian).
	TagLong64U Tag = 21

	// TagEnum represents an enumeration (treated as unsigned).
	// Range: 0 to 255.
	// Encoding: Tag (0x16), value (1 byte).
	TagEnum Tag = 22

	// TagFloat32 represents a 32-bit floating-point number (IEEE 754).
	// Range: As per IEEE 754 single precision.
	// Encoding: Tag (0x17), value (4 bytes, big-endian).
	TagFloat32 Tag = 23

	// TagFloat64 represents a 64-bit floating-point number (IEEE 754).
	// Range: As per IEEE 754 double precision.
	// Encoding: Tag (0x18), value (8 bytes, big-endian).
	TagFloat64 Tag = 24

	// TagDateTime represents a date and time value (12 bytes).
	// Encoding: Tag (0x19), followed by 12 bytes as per IEC 62056-6-2 clause 4.1.6.1.
	// Fields: Year (0x0000–0xFFFE, 0xFFFF undefined), Month (1–12, 0xFD/0xFE DST, 0xFF undefined),
	// Day (1–31, 0xFD/0xFE special, 0xFF undefined), DayOfWeek (1–7, 0xFF undefined),
	// Hour (0–23, 0xFF undefined), Minute/Second (0–59, 0xFF undefined), Hundredths (0–99, 0xFF undefined),
	// Deviation (-720 to +840 minutes, 0x8000 not specified), ClockStatus (bit flags, 0xFF not specified).
	TagDateTime Tag = 25

	// TagDate represents a date value (5 bytes).
	// Encoding: Tag (0x1A), followed by 5 bytes as per IEC 62056-6-2 clause 4.1.6.1.
	// Fields: Year (0x0000–0xFFFE, 0xFFFF undefined), Month (1–12, 0xFD/0xFE DST, 0xFF undefined),
	// Day (1–31, 0xFD/0xFE special, 0xFF undefined), DayOfWeek (1–7, 0xFF undefined).
	TagDate Tag = 26

	// TagTime represents a time value (4 bytes).
	// Encoding: Tag (0x1B), followed by 4 bytes as per IEC 62056-6-2 clause 4.1.6.1.
	// Fields: Hour (0–23, 0xFF undefined), Minute/Second (0–59, 0xFF undefined), Hundredths (0–99, 0xFF undefined).
	TagTime Tag = 27

	// TagDeltaInteger represents a delta-encoded 8-bit signed integer.
	// Range: -128 to 127.
	// Encoding: Tag (0x1C), value (1 byte).
	TagDeltaInteger Tag = 28

	// TagDeltaLong represents a delta-encoded 16-bit signed integer.
	// Range: -32,768 to 32,767.
	// Encoding: Tag (0x1D), value (2 bytes, big-endian).
	TagDeltaLong Tag = 29

	// TagDeltaDoubleLong represents a delta-encoded 32-bit signed integer.
	// Range: -2,147,483,648 to 2,147,483,647.
	// Encoding: Tag (0x1E), value (4 bytes, big-endian).
	TagDeltaDoubleLong Tag = 30

	// TagDeltaUnsigned represents a delta-encoded 8-bit unsigned integer.
	// Range: 0 to 255.
	// Encoding: Tag (0x1F), value (1 byte).
	TagDeltaUnsigned Tag = 31

	// TagDeltaLongUnsigned represents a delta-encoded 16-bit unsigned integer.
	// Range: 0 to 65,535.
	// Encoding: Tag (0x20), value (2 bytes, big-endian).
	TagDeltaLongUnsigned Tag = 32

	// TagDeltaDoubleLongUnsigned represents a delta-encoded 32-bit unsigned integer.
	// Range: 0 to 4,294,967,295.
	// Encoding: Tag (0x21), value (4 bytes, big-endian).
	TagDeltaDoubleLongUnsigned Tag = 33

	// TagDontCare represents an undefined or ignored value.
	// Encoding: Tag (0xFF).
	TagDontCare Tag = 255
)
