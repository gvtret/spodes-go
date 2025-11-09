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

	t.Run("Read Access", func(t *testing.T) {
		val, err := base.GetAttribute(1)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", val)
	})

	t.Run("Read/Write Access", func(t *testing.T) {
		val, err := base.GetAttribute(2)
		assert.NoError(t, err)
		assert.Equal(t, 123, val)
	})

	t.Run("Write-Only Access", func(t *testing.T) {
		_, err := base.GetAttribute(3)
		assert.Equal(t, ErrAccessDenied, err)
	})

	t.Run("Non-existent Attribute", func(t *testing.T) {
		_, err := base.GetAttribute(4)
		assert.Equal(t, ErrAttributeNotSupported, err)
	})
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

	t.Run("Read-Only Access", func(t *testing.T) {
		err := base.SetAttribute(1, "new_value")
		assert.Equal(t, ErrAccessDenied, err)
	})

	t.Run("Read/Write Access", func(t *testing.T) {
		err := base.SetAttribute(2, 456)
		assert.NoError(t, err)
		val, _ := base.GetAttribute(2)
		assert.Equal(t, 456, val)
	})

	t.Run("Write-Only Access", func(t *testing.T) {
		err := base.SetAttribute(3, false)
		assert.NoError(t, err)
		assert.Equal(t, false, base.Attributes[3].Value)
	})

	t.Run("Non-existent Attribute", func(t *testing.T) {
		err := base.SetAttribute(4, "some_value")
		assert.Equal(t, ErrAttributeNotSupported, err)
	})

	t.Run("Invalid Value Type", func(t *testing.T) {
		err := base.SetAttribute(2, "invalid_type")
		assert.Equal(t, ErrInvalidValueType, err)
	})
}

func TestBaseImpl_GetAttributeAccess(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.2.3.4.5.6")
	base := &BaseImpl{
		ClassID:    1,
		InstanceID: *obis,
		Attributes: map[byte]AttributeDescriptor{
			1: {Access: AttributeRead},
			2: {Access: AttributeWrite},
			3: {Access: AttributeRead | AttributeWrite},
		},
	}

	assert.Equal(t, AttributeRead, base.GetAttributeAccess(1))
	assert.Equal(t, AttributeWrite, base.GetAttributeAccess(2))
	assert.Equal(t, AttributeRead|AttributeWrite, base.GetAttributeAccess(3))
	assert.Equal(t, AttributeNoAccess, base.GetAttributeAccess(4))
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

	t.Run("Successful Invocation", func(t *testing.T) {
		result, err := base.Invoke(1, []interface{}{5})
		assert.NoError(t, err)
		assert.Equal(t, 10, result)
	})

	t.Run("Non-existent Method", func(t *testing.T) {
		_, err := base.Invoke(3, []interface{}{})
		assert.Equal(t, ErrMethodNotSupported, err)
	})

	t.Run("No Access Method", func(t *testing.T) {
		_, err := base.Invoke(2, []interface{}{})
		assert.Equal(t, ErrAccessDenied, err)
	})

	t.Run("Invalid Parameters (Number)", func(t *testing.T) {
		_, err := base.Invoke(1, []interface{}{})
		assert.Equal(t, ErrInvalidParameter, err)
	})

	t.Run("Invalid Parameters (Type)", func(t *testing.T) {
		_, err := base.Invoke(1, []interface{}{"wrong_type"})
		assert.Equal(t, ErrInvalidParameter, err)
	})
}

func TestBaseImpl_GetMethodAccess(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.2.3.4.5.6")
	base := &BaseImpl{
		ClassID:    1,
		InstanceID: *obis,
		Methods: map[byte]MethodDescriptor{
			1: {Access: MethodAccessAllowed},
			2: {Access: MethodNoAccess},
		},
	}

	assert.Equal(t, MethodAccessAllowed, base.GetMethodAccess(1))
	assert.Equal(t, MethodNoAccess, base.GetMethodAccess(2))
	assert.Equal(t, MethodNoAccess, base.GetMethodAccess(3))
}
