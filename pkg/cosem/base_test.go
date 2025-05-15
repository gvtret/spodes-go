package cosem

import (
	"fmt"
	"github.com/gvtret/spodes-go/pkg/axdr"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// TestBaseImplGetClassID verifies retrieval of ClassID.
func TestBaseImplGetClassID(t *testing.T) {
	base := BaseImpl{ClassID: 42}
	assert.Equal(t, uint16(42), base.GetClassID())
}

// TestBaseImplGetInstanceID verifies retrieval of InstanceID.
func TestBaseImplGetInstanceID(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	base := BaseImpl{InstanceID: *obis}
	assert.Equal(t, *obis, base.GetInstanceID())
}

// TestBaseImplGetAttribute verifies attribute retrieval.
func TestBaseImplGetAttribute(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.1.8.0.255")
	base := BaseImpl{
		Attributes: map[byte]AttributeDescriptor{
			1: {
				Type:   reflect.TypeOf(*obis),
				Access: AttributeRead,
				Value:  *obis,
			},
			2: {
				Type:   reflect.TypeOf(uint32(0)),
				Access: AttributeWrite, // Write-only
				Value:  uint32(12345),
			},
			3: {
				Type:   reflect.TypeOf(axdr.Array{uint32(0)}),
				Access: AttributeRead | AttributeWrite,
				Value:  axdr.Array{uint32(1), uint32(2)},
			},
		},
	}

	tests := []struct {
		name        string
		attributeID byte
		expected    interface{}
		expectedErr error
	}{
		{
			name:        "ReadAccessibleAttributeObisCode",
			attributeID: 1,
			expected:    *obis,
			expectedErr: nil,
		},
		{
			name:        "ReadInaccessibleAttribute",
			attributeID: 2,
			expected:    nil,
			expectedErr: ErrAccessDenied,
		},
		{
			name:        "ReadArrayAttribute",
			attributeID: 3,
			expected:    axdr.Array{uint32(1), uint32(2)},
			expectedErr: nil,
		},
		{
			name:        "NonExistentAttribute",
			attributeID: 4,
			expected:    nil,
			expectedErr: ErrAttributeNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base.GetAttribute(tt.attributeID)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err, "Expected error for attribute access")
				assert.Nil(t, result, "Result should be nil for error cases")
				return
			}
			assert.NoError(t, err, "GetAttribute should succeed for readable attribute")
			assert.Equal(t, tt.expected, result, "Attribute value should match expected")
		})
	}
}

// TestBaseImplSetAttribute verifies attribute setting.
func TestBaseImplSetAttribute(t *testing.T) {
	// Initialize BaseImpl with attributes
	base := BaseImpl{
		Attributes: map[byte]AttributeDescriptor{
			1: {
				Type:   reflect.TypeOf(ObisCode{}),
				Access: AttributeRead, // Read-only
				Value:  ObisCode{},    // Placeholder ObisCode
			},
			2: {
				Type:   reflect.TypeOf(uint32(0)),
				Access: AttributeWrite,
				Value:  uint32(12345),
			},
			3: {
				Type:   reflect.TypeOf(axdr.Structure{}),
				Access: AttributeRead | AttributeWrite,
				Value:  axdr.Structure{uint32(42), "initial"},
			},
		},
	}

	tests := []struct {
		name        string
		attributeID byte
		value       interface{}
		expectedErr error
	}{
		{
			name:        "SetWritableAttributeUint32",
			attributeID: 2,
			value:       uint32(67890),
			expectedErr: nil,
		},
		{
			name:        "SetWritableAttributeStructure",
			attributeID: 3,
			value:       axdr.Structure{uint32(100), "updated"},
			expectedErr: nil,
		},
		{
			name:        "SetReadOnlyAttribute",
			attributeID: 1,
			value:       ObisCode{},
			expectedErr: ErrAccessDenied,
		},
		{
			name:        "SetInvalidType",
			attributeID: 2,
			value:       "wrong type",
			expectedErr: ErrInvalidValueType,
		},
		{
			name:        "SetNonExistentAttribute",
			attributeID: 4,
			value:       uint32(67890),
			expectedErr: ErrAttributeNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify access rights before setting
			access := base.GetAttributeAccess(tt.attributeID)
			if tt.expectedErr == nil {
				assert.NotEqual(t, AttributeNoAccess, access, "Attribute should have write access")
				assert.NotEqual(t, 0, access&AttributeWrite, "Attribute should have AttributeWrite")
			}

			// Attempt to set the attribute
			err := base.SetAttribute(tt.attributeID, tt.value)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err, "Expected error for attribute setting")
				return
			}
			assert.NoError(t, err, "SetAttribute should succeed for writable attribute")

			// Verify that the value was updated (for attributes with read access)
			if access&AttributeRead != 0 {
				result, err := base.GetAttribute(tt.attributeID)
				assert.NoError(t, err, "GetAttribute should succeed for readable attribute")
				assert.Equal(t, tt.value, result, "Attribute value should match the set value")
			}
		})
	}
}

// TestBaseImplInvoke verifies method invocation.
func TestBaseImplInvoke(t *testing.T) {
	// Initialize BaseImpl with methods
	base := BaseImpl{
		Methods: map[byte]MethodDescriptor{
			1: {
				Access:     MethodAccessAllowed,
				ParamTypes: []reflect.Type{reflect.TypeOf(int(0)), reflect.TypeOf(axdr.Array{})},
				ReturnType: reflect.TypeOf(""),
				Handler: func(params []interface{}) (interface{}, error) {
					return fmt.Sprintf("Processed %d, %v", params[0].(int), params[1].(axdr.Array)), nil
				},
			},
			2: {
				Access:     MethodNoAccess,
				ParamTypes: []reflect.Type{},
				ReturnType: reflect.TypeOf(""),
				Handler:    func(params []interface{}) (interface{}, error) { return "", nil },
			},
		},
	}

	tests := []struct {
		name        string
		methodID    byte
		params      []interface{}
		expected    interface{}
		expectedErr error
	}{
		{
			name:        "InvokeAccessibleMethod",
			methodID:    1,
			params:      []interface{}{42, axdr.Array{uint32(1), uint32(2)}},
			expected:    "Processed 42, [1 2]",
			expectedErr: nil,
		},
		{
			name:        "InvokeInaccessibleMethod",
			methodID:    2,
			params:      []interface{}{},
			expected:    nil,
			expectedErr: ErrAccessDenied,
		},
		{
			name:        "InvokeNonExistentMethod",
			methodID:    3,
			params:      []interface{}{},
			expected:    nil,
			expectedErr: ErrMethodNotSupported,
		},
		{
			name:        "InvokeWithWrongParamType",
			methodID:    1,
			params:      []interface{}{"wrong", axdr.Array{uint32(1)}},
			expected:    nil,
			expectedErr: ErrInvalidParameter,
		},
		{
			name:        "InvokeWithWrongParamCount",
			methodID:    1,
			params:      []interface{}{1},
			expected:    nil,
			expectedErr: ErrInvalidParameter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base.Invoke(tt.methodID, tt.params)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBaseImplGetAttributeAccess verifies attribute access rights.
func TestBaseImplGetAttributeAccess(t *testing.T) {
	base := BaseImpl{
		Attributes: map[byte]AttributeDescriptor{
			1: {Access: AttributeRead},
			2: {Access: AttributeWrite},
			3: {Access: AttributeRead | AttributeWrite},
		},
	}

	tests := []struct {
		name        string
		attributeID byte
		expected    AttributeAccess
	}{
		{
			name:        "ReadAccess",
			attributeID: 1,
			expected:    AttributeRead,
		},
		{
			name:        "WriteAccess",
			attributeID: 2,
			expected:    AttributeWrite,
		},
		{
			name:        "ReadAndWriteAccess",
			attributeID: 3,
			expected:    AttributeRead | AttributeWrite,
		},
		{
			name:        "NonExistentAttribute",
			attributeID: 4,
			expected:    AttributeNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access := base.GetAttributeAccess(tt.attributeID)
			assert.Equal(t, tt.expected, access)
		})
	}
}

// TestBaseImplGetMethodAccess verifies method access rights.
func TestBaseImplGetMethodAccess(t *testing.T) {
	base := BaseImpl{
		Methods: map[byte]MethodDescriptor{
			1: {Access: MethodAccessAllowed},
			2: {Access: MethodNoAccess},
		},
	}

	tests := []struct {
		name     string
		methodID byte
		expected MethodAccess
	}{
		{
			name:     "AccessibleMethod",
			methodID: 1,
			expected: MethodAccessAllowed,
		},
		{
			name:     "InaccessibleMethod",
			methodID: 2,
			expected: MethodNoAccess,
		},
		{
			name:     "NonExistentMethod",
			methodID: 3,
			expected: MethodNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access := base.GetMethodAccess(tt.methodID)
			assert.Equal(t, tt.expected, access)
		})
	}
}
