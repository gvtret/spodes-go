package cosem

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseImpl_GetAttribute(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.2.3.4.5.6")
	base := &BaseImpl{
		ClassID:    1,
		InstanceID: *obis,
		Attributes: map[byte]AttributeDescriptor{
			1: {
				Type:   reflect.TypeOf("test"),
				Access: AttributeRead,
				Value:  "test_value",
			},
			2: {
				Type:   reflect.TypeOf(123),
				Access: AttributeRead | AttributeWrite,
				Value:  123,
			},
			3: {
				Type:   reflect.TypeOf(true),
				Access: AttributeWrite,
				Value:  true,
			},
		},
	}

	// Test case 1: Attribute with read access
	val, err := base.GetAttribute(1)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", val)

	// Test case 2: Attribute with read/write access
	val, err = base.GetAttribute(2)
	assert.NoError(t, err)
	assert.Equal(t, 123, val)

	// Test case 3: Attribute with write-only access (should fail)
	_, err = base.GetAttribute(3)
	assert.Equal(t, ErrAccessDenied, err)

	// Test case 4: Non-existent attribute (should fail)
	_, err = base.GetAttribute(4)
	assert.Equal(t, ErrAttributeNotSupported, err)
}

func TestBaseImpl_SetAttribute(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.2.3.4.5.6")
	base := &BaseImpl{
		ClassID:    1,
		InstanceID: *obis,
		Attributes: map[byte]AttributeDescriptor{
			1: {
				Type:   reflect.TypeOf("test"),
				Access: AttributeRead,
				Value:  "test_value",
			},
			2: {
				Type:   reflect.TypeOf(123),
				Access: AttributeRead | AttributeWrite,
				Value:  123,
			},
			3: {
				Type:   reflect.TypeOf(true),
				Access: AttributeWrite,
				Value:  true,
			},
		},
	}

	// Test case 1: Attribute with read-only access (should fail)
	err := base.SetAttribute(1, "new_value")
	assert.Equal(t, ErrAccessDenied, err)

	// Test case 2: Attribute with read/write access
	err = base.SetAttribute(2, 456)
	assert.NoError(t, err)
	val, _ := base.GetAttribute(2)
	assert.Equal(t, 456, val)

	// Test case 3: Attribute with write-only access
	err = base.SetAttribute(3, false)
	assert.NoError(t, err)
	assert.Equal(t, false, base.Attributes[3].Value)

	// Test case 4: Non-existent attribute (should fail)
	err = base.SetAttribute(4, "some_value")
	assert.Equal(t, ErrAttributeNotSupported, err)

	// Test case 5: Invalid value type (should fail)
	err = base.SetAttribute(2, "invalid_type")
	assert.Equal(t, ErrInvalidValueType, err)
}

func TestBaseImpl_GetAttributeAccess(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.2.3.4.5.6")
	base := &BaseImpl{
		ClassID:    1,
		InstanceID: *obis,
		Attributes: map[byte]AttributeDescriptor{
			1: {
				Access: AttributeRead,
			},
			2: {
				Access: AttributeWrite,
			},
			3: {
				Access: AttributeRead | AttributeWrite,
			},
		},
	}

	// Test case 1: Read-only attribute
	access := base.GetAttributeAccess(1)
	assert.Equal(t, AttributeRead, access)

	// Test case 2: Write-only attribute
	access = base.GetAttributeAccess(2)
	assert.Equal(t, AttributeWrite, access)

	// Test case 3: Read/write attribute
	access = base.GetAttributeAccess(3)
	assert.Equal(t, AttributeRead|AttributeWrite, access)

	// Test case 4: Non-existent attribute
	access = base.GetAttributeAccess(4)
	assert.Equal(t, AttributeNoAccess, access)
}

func TestBaseImpl_Invoke(t *testing.T) {
	handler := func(params []interface{}) (interface{}, error) {
		if len(params) == 1 {
			if val, ok := params[0].(int); ok {
				return val * 2, nil
			}
		}
		return nil, ErrInvalidParameter
	}

	obis, _ := NewObisCodeFromString("1.2.3.4.5.6")
	base := &BaseImpl{
		ClassID:    1,
		InstanceID: *obis,
		Methods: map[byte]MethodDescriptor{
			1: {
				Access:     MethodAccessAllowed,
				ParamTypes: []reflect.Type{reflect.TypeOf(0)},
				ReturnType: reflect.TypeOf(0),
				Handler:    handler,
			},
			2: {
				Access: MethodNoAccess,
			},
		},
	}

	// Test case 1: Successful invocation
	result, err := base.Invoke(1, []interface{}{5})
	assert.NoError(t, err)
	assert.Equal(t, 10, result)

	// Test case 2: Non-existent method (should fail)
	_, err = base.Invoke(3, []interface{}{})
	assert.Equal(t, ErrMethodNotSupported, err)

	// Test case 3: Method with no access (should fail)
	_, err = base.Invoke(2, []interface{}{})
	assert.Equal(t, ErrAccessDenied, err)

	// Test case 4: Invalid parameters (wrong number)
	_, err = base.Invoke(1, []interface{}{})
	assert.Equal(t, ErrInvalidParameter, err)

	// Test case 5: Invalid parameters (wrong type)
	_, err = base.Invoke(1, []interface{}{"wrong_type"})
	assert.Equal(t, ErrInvalidParameter, err)
}

func TestBaseImpl_GetClassID(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.1.0.0.255")
	base := BaseImpl{ClassID: 8, InstanceID: *obis}
	assert.Equal(t, uint16(8), base.GetClassID())
}

func TestBaseImpl_GetInstanceID(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.1.0.0.255")
	base := BaseImpl{ClassID: 8, InstanceID: *obis}
	assert.Equal(t, *obis, base.GetInstanceID())
}

func TestBaseImpl_GetSetAttribute(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.1.0.0.255")
	base := BaseImpl{
		ClassID:    1,
		InstanceID: *obis,
		Attributes: map[byte]AttributeDescriptor{
			1: {Type: reflect.TypeOf(""), Access: AttributeRead | AttributeWrite, Value: "initial"},
		},
	}
	// Get initial value
	val, err := base.GetAttribute(1)
	assert.NoError(t, err)
	assert.Equal(t, "initial", val)

	// Set new value
	err = base.SetAttribute(1, "new")
	assert.NoError(t, err)

	// Get new value
	val, err = base.GetAttribute(1)
	assert.NoError(t, err)
	assert.Equal(t, "new", val)

	// Set value with wrong type
	err = base.SetAttribute(1, 123)
	assert.Error(t, err)
}

func TestBaseImpl_AccessLevels(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.1.0.0.255")
	base := BaseImpl{
		ClassID:    1,
		InstanceID: *obis,
		Attributes: map[byte]AttributeDescriptor{
			1: {Access: AttributeRead},
			2: {Access: AttributeWrite},
		},
		Methods: map[byte]MethodDescriptor{
			1: {Access: MethodAccessAllowed},
			2: {Access: MethodNoAccess},
		},
	}

	// Attribute access
	assert.Equal(t, AttributeRead, base.GetAttributeAccess(1))
	assert.Equal(t, AttributeWrite, base.GetAttributeAccess(2))
	assert.Equal(t, AttributeNoAccess, base.GetAttributeAccess(3))

	// Method access
	assert.Equal(t, MethodAccessAllowed, base.GetMethodAccess(1))
	assert.Equal(t, MethodNoAccess, base.GetMethodAccess(2))
	assert.Equal(t, MethodNoAccess, base.GetMethodAccess(3))
}
