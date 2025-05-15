package cosem

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gvtret/spodes-go/pkg/axdr"
	"github.com/stretchr/testify/assert"
)

// TestRegisterNewRegister verifies the creation of a Register object with various data types.
func TestRegisterNewRegister(t *testing.T) {
	tests := []struct {
		name        string
		obis        string
		value       interface{}
		scalerUnit  ScalerUnit
		expectedErr bool
	}{
		{
			name:        "ValidObisAndUint32",
			obis:        "1.0.1.8.0.255",
			value:       uint32(12345),
			scalerUnit:  ScalerUnit{Scaler: -2, Unit: 27}, // 10^-2 Wh
			expectedErr: false,
		},
		{
			name: "ValidObisAndDateTime",
			obis: "1.0.0.1.0.255",
			value: axdr.DateTime{
				Date: axdr.Date{Year: 2025, Month: 5, Day: 15, DayOfWeek: 4},
				Time: axdr.Time{Hour: 7, Minute: 32},
			},
			scalerUnit:  ScalerUnit{Scaler: 0, Unit: 255}, // Unitless
			expectedErr: false,
		},
		{
			name:        "ValidObisAndArray",
			obis:        "1.0.99.1.0.255",
			value:       axdr.Array{uint32(1), uint32(2)},
			scalerUnit:  ScalerUnit{Scaler: 0, Unit: 30}, // V
			expectedErr: false,
		},
		{
			name:        "InvalidObisCode",
			obis:        "1.2.3",
			value:       uint32(12345),
			scalerUnit:  ScalerUnit{Scaler: -2, Unit: 27},
			expectedErr: true,
		},
		{
			name:        "InvalidValueType",
			obis:        "1.0.1.8.0.255",
			value:       complex64(1 + 2i),
			scalerUnit:  ScalerUnit{Scaler: -2, Unit: 27},
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

			register, err := NewRegister(*obis, tt.value, tt.scalerUnit)
			if tt.expectedErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, register)
			assert.Equal(t, RegisterClassID, register.GetClassID())
			assert.Equal(t, tt.obis, register.GetInstanceID().String())

			// Verify logical_name
			logicalName, err := register.GetAttribute(1)
			assert.NoError(t, err)
			assert.Equal(t, *obis, logicalName)

			// Verify value
			value, err := register.GetAttribute(2)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, value)

			// Verify scaler_unit
			scalerUnit, err := register.GetAttribute(3)
			assert.NoError(t, err)
			expectedScalerUnit := axdr.Structure{tt.scalerUnit.Scaler, tt.scalerUnit.Unit}
			assert.Equal(t, expectedScalerUnit, scalerUnit)
		})
	}
}

// TestRegisterGetAttribute verifies attribute retrieval.
func TestRegisterGetAttribute(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	scalerUnit := ScalerUnit{Scaler: -2, Unit: 27} // Wh
	register, _ := NewRegister(*obis, uint32(12345), scalerUnit)

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
			expected:    uint32(12345),
			expectedErr: nil,
		},
		{
			name:        "GetScalerUnit",
			attributeID: 3,
			expected:    axdr.Structure{scalerUnit.Scaler, scalerUnit.Unit},
			expectedErr: nil,
		},
		{
			name:        "InvalidAttribute",
			attributeID: 4,
			expected:    nil,
			expectedErr: ErrAttributeNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := register.GetAttribute(tt.attributeID)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, result)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRegisterSetAttribute verifies attribute setting.
func TestRegisterSetAttribute(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	scalerUnit := ScalerUnit{Scaler: -2, Unit: 27} // Wh
	register, _ := NewRegister(*obis, uint32(12345), scalerUnit)

	tests := []struct {
		name        string
		attributeID byte
		value       interface{}
		expectedErr error
	}{
		{
			name:        "SetValidValue",
			attributeID: 2,
			value:       uint32(67890),
			expectedErr: nil,
		},
		{
			name:        "SetValidScalerUnit",
			attributeID: 3,
			value:       axdr.Structure{int8(-1), uint8(30)}, // V
			expectedErr: nil,
		},
		{
			name:        "SetInvalidValueType",
			attributeID: 2,
			value:       "wrong type",
			expectedErr: ErrInvalidValueType,
		},
		{
			name:        "SetLogicalNameNotAllowed",
			attributeID: 1,
			value:       *obis,
			expectedErr: ErrAttributeNotSupported,
		},
		{
			name:        "SetInvalidAttribute",
			attributeID: 4,
			value:       uint32(67890),
			expectedErr: ErrAttributeNotSupported,
		},
		{
			name:        "SetInvalidAXDRType",
			attributeID: 2,
			value:       complex64(1 + 2i),
			expectedErr: fmt.Errorf("invalid value type for A-XDR encoding: %v", fmt.Errorf("unsupported type: complex64")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify attribute type before setting
			if tt.attributeID == 2 || tt.attributeID == 3 {
				currentValue, err := register.GetAttribute(tt.attributeID)
				assert.NoError(t, err)
				if tt.expectedErr == nil {
					assert.Equal(t, reflect.TypeOf(currentValue), reflect.TypeOf(tt.value), "Value type should match attribute type")
				}
			}

			err := register.SetAttribute(tt.attributeID, tt.value)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)

			// Verify updated value
			if tt.attributeID == 2 || tt.attributeID == 3 {
				value, err := register.GetAttribute(tt.attributeID)
				assert.NoError(t, err)
				assert.Equal(t, tt.value, value)
			}
		})
	}
}

// TestRegisterInvoke verifies method invocation.
func TestRegisterInvoke(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	scalerUnit := ScalerUnit{Scaler: -2, Unit: 27} // Wh
	register, _ := NewRegister(*obis, uint32(12345), scalerUnit)

	tests := []struct {
		name        string
		methodID    byte
		params      []interface{}
		expected    interface{}
		expectedErr error
	}{
		{
			name:        "ResetMethod",
			methodID:    1,
			params:      []interface{}{},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name:        "InvalidMethod",
			methodID:    2,
			params:      []interface{}{},
			expected:    nil,
			expectedErr: ErrMethodNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := register.Invoke(tt.methodID, tt.params)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, result)
				return
			}
			assert.NoError(t, err)

			// Verify reset method effect
			if tt.methodID == 1 {
				value, err := register.GetAttribute(2)
				assert.NoError(t, err)
				assert.Equal(t, uint32(0), value, "Value should be reset to default (0)")
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRegisterGetAttributeAccess verifies attribute access rights.
func TestRegisterGetAttributeAccess(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	scalerUnit := ScalerUnit{Scaler: -2, Unit: 27} // Wh
	register, _ := NewRegister(*obis, uint32(12345), scalerUnit)

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
			name:        "ScalerUnitAccess",
			attributeID: 3,
			expected:    AttributeRead | AttributeWrite,
		},
		{
			name:        "InvalidAttribute",
			attributeID: 4,
			expected:    AttributeNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access := register.GetAttributeAccess(tt.attributeID)
			assert.Equal(t, tt.expected, access)
		})
	}
}

// TestRegisterGetMethodAccess verifies method access rights.
func TestRegisterGetMethodAccess(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	scalerUnit := ScalerUnit{Scaler: -2, Unit: 27} // Wh
	register, _ := NewRegister(*obis, uint32(12345), scalerUnit)

	tests := []struct {
		name     string
		methodID byte
		expected MethodAccess
	}{
		{
			name:     "ResetMethodAccess",
			methodID: 1,
			expected: MethodAccessAllowed,
		},
		{
			name:     "InvalidMethod",
			methodID: 2,
			expected: MethodNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access := register.GetMethodAccess(tt.methodID)
			assert.Equal(t, tt.expected, access)
		})
	}
}

// TestRegisterAXDRSerialization verifies A-XDR serialization and deserialization.
func TestRegisterAXDRSerialization(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	scalerUnit := ScalerUnit{Scaler: -2, Unit: 27} // Wh
	tests := []struct {
		name       string
		value      interface{}
		scalerUnit ScalerUnit
	}{
		{
			name:       "SerializeUint32",
			value:      uint32(12345),
			scalerUnit: scalerUnit,
		},
		{
			name: "SerializeDateTime",
			value: axdr.DateTime{
				Date: axdr.Date{Year: 2025, Month: 5, Day: 15, DayOfWeek: 4},
				Time: axdr.Time{Hour: 7, Minute: 32},
			},
			scalerUnit: ScalerUnit{Scaler: 0, Unit: 255}, // Unitless
		},
		{
			name:       "SerializeArray",
			value:      axdr.Array{uint32(1), uint32(2)},
			scalerUnit: ScalerUnit{Scaler: 0, Unit: 30}, // V
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			register, err := NewRegister(*obis, tt.value, tt.scalerUnit)
			assert.NoError(t, err)

			// Test value serialization
			value, err := register.GetAttribute(2)
			assert.NoError(t, err)
			encoded, err := axdr.Encode(value)
			assert.NoError(t, err)
			decoded, err := axdr.Decode(encoded)
			assert.NoError(t, err)
			assert.Equal(t, value, decoded)

			// Test scaler_unit serialization
			scalerUnit, err := register.GetAttribute(3)
			assert.NoError(t, err)
			encoded, err = axdr.Encode(scalerUnit)
			assert.NoError(t, err)
			decoded, err = axdr.Decode(encoded)
			assert.NoError(t, err)
			assert.Equal(t, scalerUnit, decoded)
		})
	}
}
