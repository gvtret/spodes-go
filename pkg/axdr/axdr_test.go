package axdr

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

// TestPrimitiveTypes tests encoding and decoding of primitive types, including null, boolean, numeric, and delta types.
func TestPrimitiveTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []byte
		wantErr  bool
	}{
		// Null
		{
			name:     "null",
			input:    nil,
			expected: []byte{0x00},
		},

		// Boolean
		{
			name:     "boolean_true",
			input:    true,
			expected: []byte{0x03, 0x01},
		},
		{
			name:     "boolean_false",
			input:    false,
			expected: []byte{0x03, 0x00},
		},

		// Integer (int8)
		{
			name:     "integer_positive",
			input:    int8(127),
			expected: []byte{0x1C, 0x7F},
		},
		{
			name:     "integer_negative",
			input:    int8(-128),
			expected: []byte{0x1C, 0x80},
		},

		// Long (int16)
		{
			name:     "long_positive",
			input:    int16(32767),
			expected: []byte{0x1D, 0x7F, 0xFF},
		},
		{
			name:     "long_negative",
			input:    int16(-32768),
			expected: []byte{0x1D, 0x80, 0x00},
		},

		// Unsigned (uint8)
		{
			name:     "unsigned_max",
			input:    uint8(255),
			expected: []byte{0x1F, 0xFF},
		},
		{
			name:     "unsigned_zero",
			input:    uint8(0),
			expected: []byte{0x1F, 0x00},
		},

		// LongUnsigned (uint16)
		{
			name:     "long_unsigned_max",
			input:    uint16(65535),
			expected: []byte{0x20, 0xFF, 0xFF},
		},
		{
			name:     "long_unsigned_zero",
			input:    uint16(0),
			expected: []byte{0x20, 0x00, 0x00},
		},

		// DoubleLong (int32)
		{
			name:     "double_long_positive",
			input:    int32(2147483647),
			expected: []byte{0x1E, 0x7F, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "double_long_negative",
			input:    int32(-2147483648),
			expected: []byte{0x1E, 0x80, 0x00, 0x00, 0x00},
		},

		// DoubleLongUnsigned (uint32)
		{
			name:     "double_long_unsigned_max",
			input:    uint32(4294967295),
			expected: []byte{0x21, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "double_long_unsigned_zero",
			input:    uint32(0),
			expected: []byte{0x21, 0x00, 0x00, 0x00, 0x00},
		},

		// Long64 (int64)
		{
			name:     "long64_positive",
			input:    int64(9223372036854775807),
			expected: []byte{0x14, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "long64_negative",
			input:    int64(-9223372036854775808),
			expected: []byte{0x14, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},

		// Long64Unsigned (uint64)
		{
			name:     "long64_unsigned_max",
			input:    uint64(18446744073709551615),
			expected: []byte{0x15, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "long64_unsigned_zero",
			input:    uint64(0),
			expected: []byte{0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},

		// Float32
		{
			name:     "float32_positive",
			input:    float32(3.14159),
			expected: []byte{0x17, 0x40, 0x49, 0x0F, 0xD0},
		},
		{
			name:     "float32_negative",
			input:    float32(-3.14159),
			expected: []byte{0x17, 0xC0, 0x49, 0x0F, 0xD0},
		},

		// Float64
		{
			name:     "float64_positive",
			input:    float64(3.141592653589793),
			expected: []byte{0x18, 0x40, 0x09, 0x21, 0xFB, 0x54, 0x44, 0x2D, 0x18},
		},
		{
			name:     "float64_negative",
			input:    float64(-3.141592653589793),
			expected: []byte{0x18, 0xC0, 0x09, 0x21, 0xFB, 0x54, 0x44, 0x2D, 0x18},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Encode(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Encode(%v) error: %v", tt.input, err)
				return
			}
			if !bytes.Equal(encoded, tt.expected) {
				t.Errorf("Encode(%v) = %x, want %x", tt.input, encoded, tt.expected)
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Errorf("Decode(%x) error: %v", encoded, err)
				return
			}
			if !reflect.DeepEqual(decoded, tt.input) {
				t.Errorf("Decode(%x) = %+v, want %+v", encoded, decoded, tt.input)
			}
		})
	}
}

// Benchmarks for complex types encoding
func BenchmarkEncodeArraySmall(b *testing.B) {
	arr := Array{int8(1), int8(2), int8(3)}
	for i := 0; i < b.N; i++ {
		Encode(arr)
	}
}

func BenchmarkEncodeArrayMedium(b *testing.B) {
	arr := make(Array, 50)
	for j := 0; j < 50; j++ {
		arr[j] = int32(j)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(arr)
	}
}

func BenchmarkEncodeArrayLarge(b *testing.B) {
	arr := make(Array, 200)
	for j := 0; j < 200; j++ {
		arr[j] = int64(j)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(arr)
	}
}

func BenchmarkEncodeStructureSmall(b *testing.B) {
	str := Structure{int8(1), "test"}
	for i := 0; i < b.N; i++ {
		Encode(str)
	}
}

func BenchmarkEncodeStructureMedium(b *testing.B) {
	str := Structure{
		int8(1), "test", true, int32(12345),
		Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
	}
	for i := 0; i < b.N; i++ {
		Encode(str)
	}
}

func BenchmarkEncodeStructureLarge(b *testing.B) {
	str := Structure{
		int8(1), "test", true, int32(12345),
		Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
		Time{Hour: 14, Minute: 8, Second: 0, Hundredths: 0},
		float64(3.14159), BitString{Bits: []byte{0xA5}, Length: 8},
		[]byte{0x01, 0x02, 0x03}, BCD{Digits: []byte{1, 2, 3, 4}},
	}
	for i := 0; i < b.N; i++ {
		Encode(str)
	}
}

func BenchmarkEncodeCompactArray(b *testing.B) {
	ca := CompactArray{
		TypeTag: TagDeltaInteger,
		Values:  []interface{}{int8(1), int8(2), int8(3), int8(4), int8(5)},
	}
	for i := 0; i < b.N; i++ {
		Encode(ca)
	}
}

// Benchmarks for complex types decoding
func BenchmarkDecodeArraySmall(b *testing.B) {
	arr := Array{int8(1), int8(2), int8(3)}
	data, _ := Encode(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeArrayMedium(b *testing.B) {
	arr := make(Array, 50)
	for j := 0; j < 50; j++ {
		arr[j] = int32(j)
	}
	data, _ := Encode(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeArrayLarge(b *testing.B) {
	arr := make(Array, 200)
	for j := 0; j < 200; j++ {
		arr[j] = int64(j)
	}
	data, _ := Encode(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeStructureSmall(b *testing.B) {
	str := Structure{int8(1), "test"}
	data, _ := Encode(str)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeStructureMedium(b *testing.B) {
	str := Structure{
		int8(1), "test", true, int32(12345),
		Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
	}
	data, _ := Encode(str)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeStructureLarge(b *testing.B) {
	str := Structure{
		int8(1), "test", true, int32(12345),
		Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
		Time{Hour: 14, Minute: 8, Second: 0, Hundredths: 0},
		float64(3.14159), BitString{Bits: []byte{0xA5}, Length: 8},
		[]byte{0x01, 0x02, 0x03}, BCD{Digits: []byte{1, 2, 3, 4}},
	}
	data, _ := Encode(str)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeCompactArray(b *testing.B) {
	ca := CompactArray{
		TypeTag: TagDeltaInteger,
		Values:  []interface{}{int8(1), int8(2), int8(3), int8(4), int8(5)},
	}
	data, _ := Encode(ca)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

// Benchmarks for primitive types encoding
func BenchmarkEncodeBool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(true)
	}
}

func BenchmarkEncodeInt8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(int8(127))
	}
}

func BenchmarkEncodeInt16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(int16(32767))
	}
}

func BenchmarkEncodeInt32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(int32(2147483647))
	}
}

func BenchmarkEncodeInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(int64(9223372036854775807))
	}
}

func BenchmarkEncodeUint8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(uint8(255))
	}
}

func BenchmarkEncodeUint16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(uint16(65535))
	}
}

func BenchmarkEncodeUint32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(uint32(4294967295))
	}
}

func BenchmarkEncodeUint64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(uint64(18446744073709551615))
	}
}

func BenchmarkEncodeFloat32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(float32(3.14159))
	}
}

func BenchmarkEncodeFloat64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(float64(3.141592653589793))
	}
}

func BenchmarkEncodeString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode("Hello, World!")
	}
}

func BenchmarkEncodeOctetString(b *testing.B) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	for i := 0; i < b.N; i++ {
		Encode(data)
	}
}

func BenchmarkEncodeDate(b *testing.B) {
	date := Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2}
	for i := 0; i < b.N; i++ {
		Encode(date)
	}
}

func BenchmarkEncodeTime(b *testing.B) {
	timeVal := Time{Hour: 14, Minute: 8, Second: 0, Hundredths: 0}
	for i := 0; i < b.N; i++ {
		Encode(timeVal)
	}
}

func BenchmarkEncodeDateTime(b *testing.B) {
	dateTime := DateTime{
		Date:        Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
		Time:        Time{Hour: 14, Minute: 8, Second: 0, Hundredths: 0},
		Deviation:   0,
		ClockStatus: 0,
	}
	for i := 0; i < b.N; i++ {
		Encode(dateTime)
	}
}

func BenchmarkEncodeBitString(b *testing.B) {
	bitString := BitString{Bits: []byte{0xA5}, Length: 8}
	for i := 0; i < b.N; i++ {
		Encode(bitString)
	}
}

func BenchmarkEncodeBCD(b *testing.B) {
	bcd := BCD{Digits: []byte{1, 2, 3, 4}}
	for i := 0; i < b.N; i++ {
		Encode(bcd)
	}
}

// Benchmarks for primitive types decoding
func BenchmarkDecodeBool(b *testing.B) {
	data, _ := Encode(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeInt8(b *testing.B) {
	data, _ := Encode(int8(127))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeInt16(b *testing.B) {
	data, _ := Encode(int16(32767))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeInt32(b *testing.B) {
	data, _ := Encode(int32(2147483647))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeInt64(b *testing.B) {
	data, _ := Encode(int64(9223372036854775807))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeUint8(b *testing.B) {
	data, _ := Encode(uint8(255))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeUint16(b *testing.B) {
	data, _ := Encode(uint16(65535))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeUint32(b *testing.B) {
	data, _ := Encode(uint32(4294967295))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeUint64(b *testing.B) {
	data, _ := Encode(uint64(18446744073709551615))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeFloat32(b *testing.B) {
	data, _ := Encode(float32(3.14159))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeFloat64(b *testing.B) {
	data, _ := Encode(float64(3.141592653589793))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeString(b *testing.B) {
	data, _ := Encode("Hello, World!")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeOctetString(b *testing.B) {
	encodedData, _ := Encode([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(encodedData)
	}
}

func BenchmarkDecodeDate(b *testing.B) {
	data, _ := Encode(Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeTime(b *testing.B) {
	data, _ := Encode(Time{Hour: 14, Minute: 8, Second: 0, Hundredths: 0})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeDateTime(b *testing.B) {
	data, _ := Encode(DateTime{
		Date:        Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
		Time:        Time{Hour: 14, Minute: 8, Second: 0, Hundredths: 0},
		Deviation:   0,
		ClockStatus: 0,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeBitString(b *testing.B) {
	data, _ := Encode(BitString{Bits: []byte{0xA5}, Length: 8})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

func BenchmarkDecodeBCD(b *testing.B) {
	data, _ := Encode(BCD{Digits: []byte{1, 2, 3, 4}})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(data)
	}
}

// TestStringTypes tests encoding and decoding of string types (octet-string, visible-string).
func TestStringTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []byte
		wantErr  bool
	}{
		// OctetString
		{
			name:     "octet_string",
			input:    []byte{0x01, 0x02, 0x03},
			expected: []byte{0x09, 0x03, 0x01, 0x02, 0x03},
		},
		{
			name:     "octet_string_empty",
			input:    []byte{},
			expected: []byte{0x09, 0x00},
		},

		// VisibleString
		{
			name:     "visible_string",
			input:    "Hello",
			expected: []byte{0x0A, 0x05, 'H', 'e', 'l', 'l', 'o'},
		},
		{
			name:     "visible_string_empty",
			input:    "",
			expected: []byte{0x0A, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Encode(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Encode(%v) error: %v", tt.input, err)
				return
			}
			if !bytes.Equal(encoded, tt.expected) {
				t.Errorf("Encode(%v) = %x, want %x", tt.input, encoded, tt.expected)
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Errorf("Decode(%x) error: %v", encoded, err)
				return
			}
			if !reflect.DeepEqual(decoded, tt.input) {
				t.Errorf("Decode(%x) = %+v, want %+v", encoded, decoded, tt.input)
			}
		})
	}
}

// TestBitStringBCD tests encoding and decoding of BitString and BCD types.
func TestBitStringBCD(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []byte
		wantErr  bool
	}{
		// BitString
		{
			name:     "bit_string_8bits",
			input:    BitString{Bits: []byte{0xA5}, Length: 8},
			expected: []byte{0x04, 0x08, 0xA5},
		},
		{
			name:     "bit_string_12bits",
			input:    BitString{Bits: []byte{0xA5, 0xF0}, Length: 12},
			expected: []byte{0x04, 0x0C, 0xA5, 0xF0},
		},
		{
			name:    "bit_string_invalid_length",
			input:   BitString{Bits: []byte{0xA5}, Length: 16},
			wantErr: true,
		},

		// BCD
		{
			name:     "bcd_1234",
			input:    BCD{Digits: []byte{1, 2, 3, 4}},
			expected: []byte{0x0D, 0x04, 0x12, 0x34},
		},
		{
			name:     "bcd_odd_digits",
			input:    BCD{Digits: []byte{1, 2, 3}},
			expected: []byte{0x0D, 0x03, 0x12, 0x30},
		},
		{
			name:    "bcd_invalid_digit",
			input:   BCD{Digits: []byte{1, 10, 3}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Encode(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Encode(%v) error: %v", tt.input, err)
				return
			}
			if !bytes.Equal(encoded, tt.expected) {
				t.Errorf("Encode(%v) = %x, want %x", tt.input, encoded, tt.expected)
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Errorf("Decode(%x) error: %v", encoded, err)
				return
			}
			if !reflect.DeepEqual(decoded, tt.input) {
				t.Errorf("Decode(%x) = %+v, want %+v", encoded, decoded, tt.input)
			}
		})
	}
}

// TestDateTimeTypes tests encoding and decoding of Date, Time, and DateTime types.
func TestDateTimeTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []byte
		wantErr  bool
	}{
		// Date
		{
			name: "date_valid",
			input: Date{
				Year:      2025,
				Month:     5,
				Day:       13,
				DayOfWeek: 2, // Tuesday
			},
			expected: []byte{0x1A, 0x07, 0xE9, 0x05, 0x0D, 0x02},
		},
		{
			name: "date_undefined",
			input: Date{
				Year:      0xFFFF,
				Month:     0xFF,
				Day:       0xFF,
				DayOfWeek: 0xFF,
			},
			expected: []byte{0x1A, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},

		// Time
		{
			name: "time_valid",
			input: Time{
				Hour:       14,
				Minute:     8,
				Second:     0,
				Hundredths: 0,
			},
			expected: []byte{0x1B, 0x0E, 0x08, 0x00, 0x00},
		},
		{
			name: "time_undefined",
			input: Time{
				Hour:       0xFF,
				Minute:     0xFF,
				Second:     0xFF,
				Hundredths: 0xFF,
			},
			expected: []byte{0x1B, 0xFF, 0xFF, 0xFF, 0xFF},
		},

		// DateTime
		{
			name: "datetime_valid",
			input: DateTime{
				Date: Date{
					Year:      2025,
					Month:     5,
					Day:       13,
					DayOfWeek: 2,
				},
				Time: Time{
					Hour:       14,
					Minute:     8,
					Second:     0,
					Hundredths: 0,
				},
				Deviation:   0,
				ClockStatus: 0,
			},
			expected: []byte{0x19, 0x07, 0xE9, 0x05, 0x0D, 0x02, 0x0E, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "datetime_undefined",
			input: DateTime{
				Date: Date{
					Year:      0xFFFF,
					Month:     0xFF,
					Day:       0xFF,
					DayOfWeek: 0xFF,
				},
				Time: Time{
					Hour:       0xFF,
					Minute:     0xFF,
					Second:     0xFF,
					Hundredths: 0xFF,
				},
				Deviation:   -32768,
				ClockStatus: 0xFF,
			},
			expected: []byte{0x19, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Encode(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Encode(%v) error: %v", tt.input, err)
				return
			}
			if !bytes.Equal(encoded, tt.expected) {
				t.Errorf("Encode(%v) = %x, want %x", tt.input, encoded, tt.expected)
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Errorf("Decode(%x) error: %v", encoded, err)
				return
			}
			if !reflect.DeepEqual(decoded, tt.input) {
				t.Errorf("Decode(%x) = %+v, want %+v", encoded, decoded, tt.input)
			}
		})
	}
}

// TestComplexTypes tests encoding and decoding of complex types (Array, Structure, CompactArray), including nested cases.
func TestComplexTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []byte
		wantErr  bool
	}{
		// Array (homogeneous)
		{
			name:     "array_integers",
			input:    Array{int8(1), int8(2), int8(3)},
			expected: []byte{0x01, 0x03, 0x1C, 0x01, 0x1C, 0x02, 0x1C, 0x03},
		},
		{
			name:    "array_too_long",
			input:   make([]interface{}, 256),
			wantErr: true,
		},

		// Array (heterogeneous)
		{
			name: "array_mixed",
			input: Array{
				int8(1),
				"test",
				Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
			},
			expected: []byte{
				0x01, 0x03, // Array, length 3
				0x1C, 0x01, // int8
				0x0A, 0x04, 't', 'e', 's', 't', // visible-string
				0x1A, 0x07, 0xE9, 0x05, 0x0D, 0x02, // date
			},
		},

		// Structure
		{
			name: "structure_simple",
			input: Structure{
				int8(1),
				"test",
			},
			expected: []byte{
				0x02, 0x02, // Structure, length 2
				0x1C, 0x01, // int8
				0x0A, 0x04, 't', 'e', 's', 't', // visible-string
			},
		},
		{
			name: "structure_nested",
			input: Structure{
				int8(1),
				Array{int8(2), int8(3)}, // Nested array
				DateTime{
					Date: Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
					Time: Time{Hour: 14, Minute: 8},
				},
			},
			expected: []byte{
				0x02, 0x03, // Structure, length 3
				0x1C, 0x01, // int8
				0x01, 0x02, 0x1C, 0x02, 0x1C, 0x03, // array
				0x19, 0x07, 0xE9, 0x05, 0x0D, 0x02, 0x0E, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // datetime
			},
		},

		// CompactArray
		{
			name: "compact_array_integers",
			input: CompactArray{
				TypeTag: TagDeltaInteger,
				Values:  []interface{}{int8(1), int8(2), int8(3)},
			},
			expected: []byte{0x13, 0x03, 0x1C, 0x01, 0x02, 0x03},
		},
		{
			name: "compact_array_dates",
			input: CompactArray{
				TypeTag: TagDate,
				Values: []interface{}{
					Date{Year: 2025, Month: 5, Day: 13, DayOfWeek: 2},
					Date{Year: 2025, Month: 5, Day: 14, DayOfWeek: 3},
				},
			},
			expected: []byte{
				0x13, 0x02, 0x1A, // CompactArray, length 2, TagDate
				0x07, 0xE9, 0x05, 0x0D, 0x02, // Date 1
				0x07, 0xE9, 0x05, 0x0E, 0x03, // Date 2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Encode(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Encode(%v) error: %v", tt.input, err)
				return
			}
			if !bytes.Equal(encoded, tt.expected) {
				t.Errorf("Encode(%v) = %x, want %x", tt.input, encoded, tt.expected)
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Errorf("Decode(%x) error: %v", encoded, err)
				return
			}
			if !reflect.DeepEqual(decoded, tt.input) {
				t.Errorf("Decode(%x) = %+v, want %+v", encoded, decoded, tt.input)
			}
		})
	}
}

// TestDateTimeRoundTrip tests DateTime conversion to/from time.Time, including undefined values.
func TestDateTimeRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		input     time.Time
		expectDST bool
	}{
		{
			name:      "valid_datetime",
			input:     time.Date(2025, 5, 13, 14, 8, 0, 0, time.UTC),
			expectDST: false,
		},
		{
			name:      "with_dst",
			input:     time.Date(2025, 5, 13, 14, 8, 0, 0, time.FixedZone("DST", 3600)),
			expectDST: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to DateTime and back
			dt := FromTime(tt.input, tt.expectDST)
			roundTripTime, err := dt.ToTime()
			if err != nil {
				t.Errorf("ToTime error: %v", err)
				return
			}

			// Check time equality with zone awareness
			if !tt.input.Equal(roundTripTime) {
				t.Errorf("Times not equal:\nOriginal: %v\nRoundTrip: %v", tt.input, roundTripTime)
			}

			// Verify timezone offsets match
			_, origOffset := tt.input.Zone()
			_, rtOffset := roundTripTime.Zone()
			if origOffset != rtOffset {
				t.Errorf("Time zone offsets differ:\nOriginal: %d\nRoundTrip: %d", origOffset, rtOffset)
			}

			// Verify DST status matches
			if tt.input.IsDST() != roundTripTime.IsDST() {
				t.Errorf("DST mismatch:\nOriginal: %t\nRoundTrip: %t", tt.input.IsDST(), roundTripTime.IsDST())
			}

			// Additional debug output
			t.Logf("Original: %v (offset=%d, DST=%t)", tt.input, origOffset, tt.input.IsDST())
			t.Logf("Encoded: %v", dt)
			t.Logf("RoundTrip: %v (offset=%d, DST=%t)", roundTripTime, rtOffset, roundTripTime.IsDST())
		})
	}
}

// TestErrorCases tests error handling for invalid inputs.
func TestErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "unsupported_type",
			input: complex64(1 + 2i),
		},
		{
			name:  "string_too_long",
			input: string(make([]byte, 256)),
		},
		{
			name:  "octet_string_too_long",
			input: make([]byte, 256),
		},
		{
			name:  "invalid_time",
			input: Time{Hour: 24},
		},
		{
			name:  "invalid_datetime",
			input: DateTime{Deviation: -721},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Encode(tt.input)
			if err == nil {
				t.Errorf("Encode(%v) expected error, got nil", tt.input)
			}
		})
	}
}
