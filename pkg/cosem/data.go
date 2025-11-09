package cosem

import (
	"fmt"
	"reflect"

	"github.com/gvtret/spodes-go/pkg/axdr"
)

// DataClassID is the class ID for the "Data" interface class as defined in IEC 62056-6-2.
const DataClassID uint16 = 1

// DataVersion is the version of the "Data" interface class.
const DataVersion byte = 0

// Data represents the COSEM "Data" interface class (Class ID: 1, Version: 0).
// It stores a single value of any A-XDR supported type, identified by a logical name (OBIS code).
type Data struct {
	BaseImpl
}

// NewData creates a new instance of the "Data" interface class with the specified OBIS code and value.
// The value must be a type supported by the A-XDR encoder/decoder.
func NewData(obis ObisCode, value interface{}) (*Data, error) {
	// Verify the OBIS code.
	if _, err := NewObisCodeFromString(obis.String()); err != nil {
		return nil, fmt.Errorf("invalid OBIS code: %v", err)
	}

	// Verify the value type is supported by A-XDR by attempting to encode it.
	if _, err := axdr.Encode(value); err != nil {
		return nil, fmt.Errorf("invalid value type for A-XDR encoding: %v", err)
	}

	// Define attributes.
	attributes := map[byte]AttributeDescriptor{
		1: { // logical_name
			Type:   reflect.TypeOf(ObisCode{}),
			Access: AttributeRead,
			Value:  obis,
		},
		2: { // value
			Type:   reflect.TypeOf(value),
			Access: AttributeRead | AttributeWrite, // Configurator can write, others read.
			Value:  value,
		},
	}

	// The "Data" class has no methods.
	methods := map[byte]MethodDescriptor{}

	return &Data{
		BaseImpl: BaseImpl{
			ClassID:    DataClassID,
			InstanceID: obis,
			Attributes: attributes,
			Methods:    methods,
		},
	}, nil
}

// GetAttribute retrieves the value of the specified attribute.
// Supported attribute IDs:
// - 1: logical_name (ObisCode)
// - 2: value (any A-XDR supported type)
func (d *Data) GetAttribute(attributeID byte) (interface{}, error) {
	return d.BaseImpl.GetAttribute(attributeID)
}

// SetAttribute sets the value of the specified attribute.
// Only attribute 2 (value) can be set, and only by clients with write access (e.g., Configurator).
// The value type must match the existing value type.
func (d *Data) SetAttribute(attributeID byte, value interface{}) error {
	if attributeID != 2 {
		return ErrAttributeNotSupported
	}

	// Verify the new value can be encoded in A-XDR.
	if _, err := axdr.Encode(value); err != nil {
		return fmt.Errorf("invalid value type for A-XDR encoding: %v", err)
	}

	return d.BaseImpl.SetAttribute(attributeID, value)
}

// Invoke is not supported as the "Data" class has no methods.
func (d *Data) Invoke(methodID byte, parameters []interface{}) (interface{}, error) {
	return nil, ErrMethodNotSupported
}

// GetAttributeAccess returns the access level for the specified attribute.
func (d *Data) GetAttributeAccess(attributeID byte) AttributeAccess {
	return d.BaseImpl.GetAttributeAccess(attributeID)
}

// GetMethodAccess returns the access level for the specified method.
// Always returns MethodNoAccess as the "Data" class has no methods.
func (d *Data) GetMethodAccess(methodID byte) MethodAccess {
	return d.BaseImpl.GetMethodAccess(methodID)
}
