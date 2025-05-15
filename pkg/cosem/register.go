package cosem

import (
	"fmt"
	"reflect"

	"github.com/gvtret/spodes-go/pkg/axdr"
)

// RegisterClassID is the class ID for the "Register" interface class as defined in IEC 62056-6-2.
const RegisterClassID uint16 = 3

// RegisterVersion is the version of the "Register" interface class.
const RegisterVersion byte = 0

// ScalerUnit represents the scaler and unit structure for the Register class.
type ScalerUnit struct {
	Scaler int8  // Scaling factor (e.g., -2 for 10^-2)
	Unit   uint8 // Unit code (e.g., 27 for Wh, per IEC 62056-6-2)
}

// Register represents the COSEM "Register" interface class (Class ID: 3, Version: 0).
// It stores a process value with a scaler and unit, identified by a logical name (OBIS code).
type Register struct {
	BaseImpl
}

// NewRegister creates a new instance of the "Register" interface class.
// Parameters:
// - obis: OBIS code for logical_name (6-byte octet-string).
// - value: Process value (any A-XDR supported type, e.g., uint32, float32).
// - scalerUnit: Scaler and unit structure (scaler as int8, unit as uint8).
func NewRegister(obis ObisCode, value interface{}, scalerUnit ScalerUnit) (*Register, error) {
	// Verify the OBIS code
	if _, err := NewObisCodeFromString(obis.String()); err != nil {
		return nil, fmt.Errorf("invalid OBIS code: %v", err)
	}

	// Verify the value type is supported by A-XDR
	if _, err := axdr.Encode(value); err != nil {
		return nil, fmt.Errorf("invalid value type for A-XDR encoding: %v", err)
	}

	// Verify the scaler_unit type is supported by A-XDR
	scalerUnitStruct := axdr.Structure{scalerUnit.Scaler, scalerUnit.Unit}
	if _, err := axdr.Encode(scalerUnitStruct); err != nil {
		return nil, fmt.Errorf("invalid scaler_unit type for A-XDR encoding: %v", err)
	}

	// Define attributes
	attributes := map[byte]AttributeDescriptor{
		1: { // logical_name
			Type:   reflect.TypeOf(ObisCode{}),
			Access: AttributeRead,
			Value:  obis,
		},
		2: { // value
			Type:   reflect.TypeOf(value),
			Access: AttributeRead | AttributeWrite, // Configurator can write
			Value:  value,
		},
		3: { // scaler_unit
			Type:   reflect.TypeOf(scalerUnitStruct),
			Access: AttributeRead | AttributeWrite, // Configurator can write
			Value:  scalerUnitStruct,
		},
	}

	// Define methods
	methods := map[byte]MethodDescriptor{
		1: { // reset
			Access:     MethodAccessAllowed,
			ParamTypes: []reflect.Type{},
			ReturnType: nil,
			Handler: func(params []interface{}) (interface{}, error) {
				// Reset value to default (e.g., 0 or empty for the type)
				defaultValue := reflect.Zero(reflect.TypeOf(value)).Interface()
				attributes[2] = AttributeDescriptor{
					Type:   reflect.TypeOf(value),
					Access: AttributeRead | AttributeWrite,
					Value:  defaultValue,
				}
				return nil, nil
			},
		},
	}

	return &Register{
		BaseImpl: BaseImpl{
			ClassID:    RegisterClassID,
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
// - 3: scaler_unit (axdr.Structure{scaler, unit})
func (r *Register) GetAttribute(attributeID byte) (interface{}, error) {
	return r.BaseImpl.GetAttribute(attributeID)
}

// SetAttribute sets the value of the specified attribute.
// Supported attribute IDs:
// - 2: value (must match existing type)
// - 3: scaler_unit (must be axdr.Structure{int8, uint8})
func (r *Register) SetAttribute(attributeID byte, value interface{}) error {
	if attributeID != 2 && attributeID != 3 {
		return ErrAttributeNotSupported
	}

	// Verify A-XDR encoding compatibility
	if _, err := axdr.Encode(value); err != nil {
		return fmt.Errorf("invalid value type for A-XDR encoding: %v", err)
	}

	return r.BaseImpl.SetAttribute(attributeID, value)
}

// Invoke executes the specified method.
// Supported method IDs:
// - 1: reset (resets value to default)
func (r *Register) Invoke(methodID byte, parameters []interface{}) (interface{}, error) {
	return r.BaseImpl.Invoke(methodID, parameters)
}

// GetAttributeAccess returns the access level for the specified attribute.
func (r *Register) GetAttributeAccess(attributeID byte) AttributeAccess {
	return r.BaseImpl.GetAttributeAccess(attributeID)
}

// GetMethodAccess returns the access level for the specified method.
func (r *Register) GetMethodAccess(methodID byte) MethodAccess {
	return r.BaseImpl.GetMethodAccess(methodID)
}
