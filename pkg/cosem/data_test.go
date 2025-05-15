package cosem

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gvtret/spodes-go/pkg/axdr"
	"github.com/stretchr/testify/assert"
)

// TestDataNewData verifies the creation of a Data object with various data types.
func TestDataNewData(t *testing.T) {
	tests := []struct {
		name        string
		obis        string
		value       interface{}
		expectedErr bool
	}{
		{
			name:        "ValidObisAndUint32",
			obis:        "1.0.1.8.0.255",
			value:       uint32(12345),
			expectedErr: false,
		},
		{
			name:        "ValidObisAndString",
			obis:        "0.0.96.1.0.255",
			value:       "test",
			expectedErr: false,
		},
		{
			name:        "ValidObisAndDateTime",
			obis:        "1.0.0.1.0.255",
			value: axdr.DateTime{
				Date: axdr.Date{
					Year:      2025,
					Month:     5,
					Day:       15,
					DayOfWeek: 4, // Thursday
				},
				Time: axdr.Time{
					Hour:       7,
					Minute:     32,
					Second:     0,
					Hundredths: 0,
				},
				Deviation:   -180, // Moscow offset
				ClockStatus: 0x80, // DST active
			},
			expectedErr: false,
		},
		{
			name:        "ValidObisAndArray",
			obis:        "1.0.99.1.0.255",
			value:       axdr.Array{uint32(1), uint32(2), uint32(3)},
			expectedErr: false,
		},
		{
			name:        "ValidObisAndStructure",
			obis:        "1.0.99.2.0.255",
			value:       axdr.Structure{uint32(42), "test", true},
			expectedErr: false,
		},
		{
			name:        "ValidObisAndCompactArray",
			obis:        "1.0.99.3.0.255",
			value: axdr.CompactArray{
				TypeTag: axdr.TagLongUnsigned,
				Values:  []interface{}{uint16(100), uint16(200)},
			},
			expectedErr: false,
		},
		{
			name:        "ValidObisAndBitString",
			obis:        "1.0.96.5.0.255",
			value: axdr.BitString{
				Bits:   []byte{0xF0}, // 11110000
				Length: 4,
			},
			expectedErr: false,
		},
		{
			name:        "ValidObisAndBCD",
			obis:        "1.0.96.6.0.255",
			value: axdr.BCD{
				Digits: []byte{1, 2, 3, 4},
			},
			expectedErr: false,
		},
		{
			name:        "InvalidObisCode",
			obis:        "1.2.3", // Invalid format
			value:       uint32(12345),
			expectedErr: true,
		},
		{
			name:        "InvalidValueType",
			obis:        "1.0.1.8.0.255",
			value:       complex64(1 + 2i), // Type not supported by A-XDR
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obis, err := NewObisCodeFromString(tt.obis)
			if tt.expectedErr && err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			data, err := NewData(*obis, tt.value)
			if tt.expectedErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, data)
			assert.Equal(t, DataClassID, data.GetClassID())
			assert.Equal(t, tt.obis, data.GetInstanceID().String())

			// Verify logical_name attribute
			logicalName, err := data.GetAttribute(1)
			assert.NoError(t, err)
			assert.Equal(t, *obis, logicalName)

			// Verify value attribute
			value, err := data.GetAttribute(2)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, value)
		})
	}
}

// TestDataGetAttribute verifies the retrieval of attributes.
func TestDataGetAttribute(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	data, _ := NewData(*obis, axdr.Array{uint32(1), uint32(2)})

	tests := []struct {
		name        string
		attributeID byte
		expected    interface{}
		expectedErr error
	}{
		{
			name:        "GetLogicalName",
			attributeID: 1,
			expected:    *obis,
			expectedErr: nil,
		},
		{
			name:        "GetValue",
			attributeID: 2,
			expected:    axdr.Array{uint32(1), uint32(2)},
			expectedErr: nil,
		},
		{
			name:        "InvalidAttribute",
			attributeID: 3,
			expected:    nil,
			expectedErr: ErrAttributeNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := data.GetAttribute(tt.attributeID)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDataSetAttribute verifies the setting of attributes.
func TestDataSetAttribute(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")

	// Initialize Data objects for different types
	arrayData, _ := NewData(*obis, axdr.Array{uint32(1), uint32(2)})
	dateTimeData, _ := NewData(*obis, axdr.DateTime{
		Date: axdr.Date{Year: 2025, Month: 5, Day: 15, DayOfWeek: 4},
		Time: axdr.Time{Hour: 7, Minute: 32},
	})

	tests := []struct {
		name        string
		data        *Data
		attributeID byte
		value       interface{}
		expectedErr error
	}{
		{
			name:        "SetValidArray",
			data:        arrayData,
			attributeID: 2,
			value:       axdr.Array{uint32(3), uint32(4)},
			expectedErr: nil,
		},
		{
			name:        "SetValidDateTime",
			data:        dateTimeData,
			attributeID: 2,
			value: axdr.DateTime{
				Date: axdr.Date{Year: 2025, Month: 5, Day: 15, DayOfWeek: 4},
				Time: axdr.Time{Hour: 7, Minute: 32},
			},
			expectedErr: nil,
		},
		{
			name:        "SetInvalidValueType",
			data:        arrayData,
			attributeID: 2,
			value:       "wrong type", // Type does not match Array
			expectedErr: ErrInvalidValueType,
		},
		{
			name:        "SetLogicalNameNotAllowed",
			data:        arrayData,
			attributeID: 1,
			value:       *obis,
			expectedErr: ErrAttributeNotSupported,
		},
		{
			name:        "SetInvalidAttribute",
			data:        arrayData,
			attributeID: 3,
			value:       axdr.Array{uint32(5)},
			expectedErr: ErrAttributeNotSupported,
		},
		{
			name:        "SetInvalidAXDRType",
			data:        arrayData,
			attributeID: 2,
			value:       complex64(1 + 2i),
			expectedErr: fmt.Errorf("invalid value type for A-XDR encoding: %w", fmt.Errorf("unsupported type: complex64")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify attribute type before setting
			currentValue, err := tt.data.GetAttribute(tt.attributeID)
			if tt.attributeID == 2 && tt.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, reflect.TypeOf(currentValue), reflect.TypeOf(tt.value), "Value type should match attribute type")
			}

			err = tt.data.SetAttribute(tt.attributeID, tt.value)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "Expected error for attribute setting")
				return
			}
			assert.NoError(t, err, "SetAttribute should succeed for valid value")

			// Verify that the value was updated
			if tt.attributeID == 2 {
				value, err := tt.data.GetAttribute(tt.attributeID)
				assert.NoError(t, err, "GetAttribute should succeed after setting")
				assert.Equal(t, tt.value, value, "Attribute value should match the set value")
			}
		})
	}
}

// TestDataInvoke verifies that method invocation is not supported.
func TestDataInvoke(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	data, _ := NewData(*obis, uint32(12345))

	result, err := data.Invoke(1, []interface{}{})
	assert.Equal(t, ErrMethodNotSupported, err)
	assert.Nil(t, result)
}

// TestDataGetAttributeAccess verifies attribute access rights.
func TestDataGetAttributeAccess(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	data, _ := NewData(*obis, axdr.CompactArray{TypeTag: axdr.TagLongUnsigned, Values: []interface{}{uint16(100)}})

	tests := []struct {
		name        string
		attributeID byte
		expected    AttributeAccess
	}{
		{
			name:        "LogicalNameAccess",
			attributeID: 1,
			expected:    AttributeRead,
		},
		{
			name:        "ValueAccess",
			attributeID: 2,
			expected:    AttributeRead | AttributeWrite,
		},
		{
			name:        "InvalidAttribute",
			attributeID: 3,
			expected:    AttributeNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access := data.GetAttributeAccess(tt.attributeID)
			assert.Equal(t, tt.expected, access)
		})
	}
}

// TestDataGetMethodAccess verifies method access rights.
func TestDataGetMethodAccess(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	data, _ := NewData(*obis, uint32(12345))

	access := data.GetMethodAccess(1)
	assert.Equal(t, MethodNoAccess, access)
}

// TestDataAXDRSerialization verifies A-XDR serialization and deserialization.
func TestDataAXDRSerialization(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "SerializeUint32",
			value: uint32(12345),
		},
		{
			name: "SerializeDateTime",
			value: axdr.DateTime{
				Date: axdr.Date{Year: 2025, Month: 5, Day: 15, DayOfWeek: 4},
				Time: axdr.Time{Hour: 7, Minute: 32},
			},
		},
		{
			name:  "SerializeArray",
			value: axdr.Array{uint32(1), uint32(2)},
		},
		{
			name:  "SerializeStructure",
			value: axdr.Structure{uint32(42), "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := NewData(*obis, tt.value)
			assert.NoError(t, err)

			// Retrieve value
			value, err := data.GetAttribute(2)
			assert.NoError(t, err)

			// Serialize to A-XDR
			encoded, err := axdr.Encode(value)
			assert.NoError(t, err)

			// Deserialize back
			decoded, err := axdr.Decode(encoded)
			assert.NoError(t, err)
			assert.Equal(t, value, decoded)
		})
	}
}