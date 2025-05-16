package axdr

import (
	"errors"
	"fmt"
	"time"
)

// type Boolean - Tag[3]
type Boolean bool

// type DoubleLong - Tag[5]
type DoubleLong int32

// type DoubleLongUnsigned - Tag[36]
type DoubleLongUnsigned uint32

// type OctetString - Tag[9]
type OctetString []byte

// type VisibleString - Tag[10]
type VisibleString []byte

// type Utf8String - Tag[12]
type Utf8String []byte

// type Integer - Tag[15]
type Integer int8

// type Long - Tag[16]
type Long int16

// type Unsigned - Tag[17]
type Unsigned uint8

// type LongUnsigned - Tag[18]
type LongUnsigned uint16

// type Long64 - Tag[20]
type Long64 int64

// type Long64Unsigned - Tag[21]
type Long64Unsigned uint64

// type Enum - Tag[22]
type Enum byte

// type Float32 - Tag[23]
type Float32 float32

// type Float64 - Tag[24]
type Float64 float64

// Array represents an A-XDR Array (homogeneous elements)
type Array []interface{}

// Structure represents an A-XDR Structure (heterogeneous fields)
type Structure []interface{}

// Date represents an A-XDR Date (5 bytes) as defined in IEC 62056-6-2 clause 4.1.6.1.
// It encodes a date with year, month, day, and day of week, supporting special values for undefined or DST transitions.
type Date struct {
	Year      uint16 // Year: 0x0000-0xFFFE (valid years), 0xFFFF (undefined).
	Month     byte   // Month: 1-12 (January-December), 0xFD (end of DST), 0xFE (start of DST), 0xFF (undefined).
	Day       byte   // Day of month: 1-31, 0xFD (second-to-last day), 0xFE (last day), 0xFF (undefined).
	DayOfWeek byte   // Day of week: 1-7 (1=Monday, 7=Sunday), 0xFF (undefined).
}

// Validate ensures Date fields are within the valid ranges specified in IEC 62056-6-2 clause 4.1.6.1.
// Returns an error if any field is invalid.
func (d Date) Validate() error {
	if d.Year > 0xFFFE && d.Year != 0xFFFF {
		return fmt.Errorf("invalid year: %d, must be 0x0000-0xFFFE or 0xFFFF", d.Year)
	}
	if d.Month > 12 && d.Month != 0xFD && d.Month != 0xFE && d.Month != 0xFF {
		return fmt.Errorf("invalid month: %d, must be 1-12, 0xFD, 0xFE, or 0xFF", d.Month)
	}
	if d.Day > 31 && d.Day != 0xFD && d.Day != 0xFE && d.Day != 0xFF {
		return fmt.Errorf("invalid day: %d, must be 1-31, 0xFD, 0xFE, or 0xFF", d.Day)
	}
	if d.DayOfWeek > 7 && d.DayOfWeek != 0xFF {
		return fmt.Errorf("invalid day of week: %d, must be 1-7 or 0xFF", d.DayOfWeek)
	}
	return nil
}

// Time represents an A-XDR Time (4 bytes) as defined in IEC 62056-6-2 clause 4.1.6.1.
// It encodes a time with hour, minute, second, and hundredths of a second, supporting undefined values.
type Time struct {
	Hour       byte // Hour: 0-23, 0xFF (undefined).
	Minute     byte // Minute: 0-59, 0xFF (undefined).
	Second     byte // Second: 0-59, 0xFF (undefined).
	Hundredths byte // Hundredths of a second: 0-99, 0xFF (undefined).
}

// Validate ensures Time fields are within the valid ranges specified in IEC 62056-6-2 clause 4.1.6.1.
// Returns an error if any field is invalid.
func (t Time) Validate() error {
	if t.Hour > 23 && t.Hour != 0xFF {
		return fmt.Errorf("invalid hour: %d, must be 0-23 or 0xFF", t.Hour)
	}
	if t.Minute > 59 && t.Minute != 0xFF {
		return fmt.Errorf("invalid minute: %d, must be 0-59 or 0xFF", t.Minute)
	}
	if t.Second > 59 && t.Second != 0xFF {
		return fmt.Errorf("invalid second: %d, must be 0-59 or 0xFF", t.Second)
	}
	if t.Hundredths > 99 && t.Hundredths != 0xFF {
		return fmt.Errorf("invalid hundredths: %d, must be 0-99 or 0xFF", t.Hundredths)
	}
	return nil
}

// DateTime represents an A-XDR DateTime (12 bytes) as defined in IEC 62056-6-2 clause 4.1.6.1.
// It combines Date and Time with deviation from UTC and clock status, supporting special values.
type DateTime struct {
	Date        Date  // Date component (year, month, day, day of week).
	Time        Time  // Time component (hour, minute, second, hundredths).
	Deviation   int16 // Deviation from UTC: -720 to +840 minutes, -32768 (0x8000) (not specified).
	ClockStatus byte  // Clock status: Bit 0 (invalid), Bit 1 (doubtful), Bit 2 (different base), Bit 3 (invalid status), Bit 7 (DST), 0xFF (not specified).
}

// Validate ensures DateTime fields are within the valid ranges specified in IEC 62056-6-2 clause 4.1.6.1.
// Returns an error if any field is invalid.
func (dt DateTime) Validate() error {
	if err := dt.Date.Validate(); err != nil {
		return err
	}
	if err := dt.Time.Validate(); err != nil {
		return err
	}
	if dt.Deviation != -32768 && (dt.Deviation < -720 || dt.Deviation > 840) {
		return fmt.Errorf("invalid deviation: %d, must be -720 to +840 or -32768 (0x8000)", dt.Deviation)
	}
	return nil
}

// FromTime converts a Go time.Time to a DateTime, mapping fields according to IEC 62056-6-2 clause 4.1.6.1.
// Special values (e.g., 0xFF) are used for undefined or out-of-range fields.
func FromTime(t time.Time, isDST bool) DateTime {
	year := uint16(t.Year())
	if year > 0xFFFE {
		year = 0xFFFF
	}
	month := byte(t.Month())
	if month < 1 || month > 12 {
		month = 0xFF
	}
	day := byte(t.Day())
	if day < 1 || day > 31 {
		day = 0xFF
	}
	dayOfWeek := byte(t.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7 // Sunday
	}
	hour := byte(t.Hour())
	if hour > 23 {
		hour = 0xFF
	}
	minute := byte(t.Minute())
	if minute > 59 {
		minute = 0xFF
	}
	second := byte(t.Second())
	if second > 59 {
		second = 0xFF
	}
	hundredths := byte(t.Nanosecond() / 1e7)
	if hundredths > 99 {
		hundredths = 0xFF
	}
	_, offset := t.Zone()
	deviation := int16(offset / 60)
	if isDST {
		deviation -= 60
	}
	if deviation < -720 || deviation > 840 {
		deviation = -32768
	}
	clockStatus := byte(0)
	if isDST {
		clockStatus |= 0x80
	}
	return DateTime{
		Date: Date{
			Year:      year,
			Month:     month,
			Day:       day,
			DayOfWeek: dayOfWeek,
		},
		Time: Time{
			Hour:       hour,
			Minute:     minute,
			Second:     second,
			Hundredths: hundredths,
		},
		Deviation:   deviation,
		ClockStatus: clockStatus,
	}
}

// ToTime converts a DateTime to a Go time.Time, handling undefined values as specified in IEC 62056-6-2 clause 4.1.6.1.
// Undefined fields (0xFF) are set to default values (e.g., 1 for month/day). Returns an error if the day of week is invalid.
func (dt DateTime) ToTime() (time.Time, error) {
	year := int(dt.Date.Year)
	if year == 0xFFFF {
		year = 0
	}
	month := int(dt.Date.Month)
	if month == 0xFF || month == 0xFD || month == 0xFE {
		month = 1
	}
	day := int(dt.Date.Day)
	if day == 0xFF || day == 0xFD || day == 0xFE {
		day = 1
	}
	hour := int(dt.Time.Hour)
	if hour == 0xFF {
		hour = 0
	}
	minute := int(dt.Time.Minute)
	if minute == 0xFF {
		minute = 0
	}
	second := int(dt.Time.Second)
	if second == 0xFF {
		second = 0
	}
	hundredths := int(dt.Time.Hundredths)
	if hundredths == 0xFF {
		hundredths = 0
	}
	totalOffset := 0
	if dt.Deviation != -32768 {
		totalOffset = int(dt.Deviation) * 60
		if dt.ClockStatus&0x80 != 0 {
			totalOffset += 3600
		}
	}

	loc := time.FixedZone("", totalOffset)
	t := time.Date(year, time.Month(month), day, hour, minute, second, hundredths*1e7, loc)

	if dt.Date.DayOfWeek != 0xFF && dt.Date.DayOfWeek != byte(t.Weekday()) && t.Weekday() != 0 {
		return time.Time{}, errors.New("invalid day of week")
	}
	return t, nil
}

// String returns a string representation of the DateTime.
//
// The returned string is in the format "YYYY-MM-DD HH:MM:SS (offset=O, DST=D)",
// where YYYY is the year, MM is the month, DD is the day, HH is the hour, MM is the minute,
// SS is the second, O is the offset in minutes, and D is a boolean indicating whether daylight saving time is in effect.
func (dt DateTime) String() string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d (offset=%d, DST=%v)",
		dt.Date.Year, dt.Date.Month, dt.Date.Day,
		dt.Time.Hour, dt.Time.Minute, dt.Time.Second,
		dt.Deviation*60, dt.ClockStatus&0x80 != 0)
}

// BitString represents an A-XDR BitString (TagBitString) as defined in IEC 62056-6-2.
// It encodes a sequence of bits, with the least significant bits used first.
// The Length field specifies the number of valid bits, and unused bits in the last byte are padded with zeros.
type BitString struct {
	Bits   []byte // Byte sequence containing the bits.
	Length uint8  // Number of valid bits (0-255).
}

// Validate ensures BitString fields are valid per IEC 62056-6-2.
// Checks that Length is within 0-255 and the byte count matches ceiling(Length/8).
func (bs BitString) Validate() error {
	expectedBytes := (bs.Length + 7) / 8
	if len(bs.Bits) != int(expectedBytes) {
		return fmt.Errorf("invalid bitstring data: %d bytes, expected %d for %d bits", len(bs.Bits), expectedBytes, bs.Length)
	}
	return nil
}

// BCD represents an A-XDR Binary-Coded Decimal (TagBCD) as defined in IEC 62056-6-2.
// It encodes a sequence of decimal digits (0-9), with each byte containing two digits (high nibble first).
type BCD struct {
	Digits []byte // Decimal digits (0-9).
}

// Validate ensures BCD fields are valid per IEC 62056-6-2.
// Checks that the number of digits is 0-255 and each digit is 0-9.
func (bcd BCD) Validate() error {
	if len(bcd.Digits) > 255 {
		return fmt.Errorf("invalid BCD length: %d, must be 0-255 digits", len(bcd.Digits))
	}
	for i, digit := range bcd.Digits {
		if digit > 9 {
			return fmt.Errorf("invalid BCD digit at index %d: %d, must be 0-9", i, digit)
		}
	}
	return nil
}
