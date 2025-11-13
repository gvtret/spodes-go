package cosem

import (
	"fmt"
	"reflect"
)

func validateCapturePeriod(value interface{}) error {
	capturePeriod := value.(uint32)
	if capturePeriod == 0 {
		return fmt.Errorf("%w: capture_period must be greater than 0", ErrInvalidParameter)
	}
	return nil
}

func validateSortMethod(value interface{}) error {
	sortMethod := value.(uint8)
	switch sortMethod {
	case 0, 1, 2, 3, 4, 5:
		return nil
	default:
		return fmt.Errorf("%w: sort_method value %d is not supported", ErrInvalidParameter, sortMethod)
	}
}

func validateSortObject(value interface{}) error {
	sortObject := value.(CosemAttributeDescriptor)

	if sortObject.ClassID == 0 && sortObject.AttributeID == 0 && sortObject.InstanceID.String() == "" {
		return nil
	}

	if sortObject.ClassID == 0 {
		return fmt.Errorf("%w: sort_object class_id must be set", ErrInvalidParameter)
	}
	if sortObject.AttributeID <= 0 {
		return fmt.Errorf("%w: sort_object attribute_id must be positive", ErrInvalidParameter)
	}
	if sortObject.InstanceID.String() == "" {
		return fmt.Errorf("%w: sort_object instance_id must be set", ErrInvalidParameter)
	}

	return nil
}

// ProfileGenericClassID is the class ID for the "Profile generic" interface class.
const ProfileGenericClassID uint16 = 7

// ProfileGenericVersion is the version of the "Profile generic" interface class.
const ProfileGenericVersion byte = 0

// ProfileGeneric represents the COSEM "Profile generic" interface class.
type ProfileGeneric struct {
	BaseImpl
}

// CaptureObjectDefinition defines a capture object for the Profile Generic class.
type CaptureObjectDefinition struct {
	ClassID     uint16
	InstanceID  ObisCode
	AttributeID uint8
	DataIndex   uint16
}

// NewProfileGeneric creates a new instance of the "Profile generic" interface class.
func NewProfileGeneric(obis ObisCode, buffer interface{}, captureObjects []CaptureObjectDefinition, capturePeriod uint32, sortMethod uint8, sortObject CosemAttributeDescriptor) (*ProfileGeneric, error) {
	attributes := map[byte]AttributeDescriptor{
		1: { // logical_name
			Type:   reflect.TypeOf(ObisCode{}),
			Access: AttributeRead,
			Value:  obis,
		},
		2: { // buffer
			Type:   reflect.TypeOf(buffer),
			Access: AttributeRead,
			Value:  buffer,
		},
		3: { // capture_objects
			Type:   reflect.TypeOf(captureObjects),
			Access: AttributeRead,
			Value:  captureObjects,
		},
		4: { // capture_period
			Type:      reflect.TypeOf(capturePeriod),
			Access:    AttributeRead | AttributeWrite,
			Value:     capturePeriod,
			Validator: validateCapturePeriod,
		},
		5: { // sort_method
			Type:      reflect.TypeOf(sortMethod),
			Access:    AttributeRead | AttributeWrite,
			Value:     sortMethod,
			Validator: validateSortMethod,
		},
		6: { // sort_object
			Type:      reflect.TypeOf(sortObject),
			Access:    AttributeRead | AttributeWrite,
			Value:     sortObject,
			Validator: validateSortObject,
		},
		7: { // entries_in_use
			Type:   reflect.TypeOf(uint32(0)),
			Access: AttributeRead,
			Value:  uint32(0),
		},
		8: { // profile_entries
			Type:   reflect.TypeOf(uint32(0)),
			Access: AttributeRead,
			Value:  uint32(0),
		},
	}

	var captureParamType reflect.Type
	if attributes[2].Type.Kind() == reflect.Slice {
		captureParamType = attributes[2].Type.Elem()
	} else {
		captureParamType = reflect.TypeOf((*interface{})(nil)).Elem()
	}

	methods := make(map[byte]MethodDescriptor)

	pg := &ProfileGeneric{}
	pg.BaseImpl = BaseImpl{
		ClassID:    ProfileGenericClassID,
		InstanceID: obis,
		Attributes: attributes,
		Methods:    methods,
	}

	methods[1] = MethodDescriptor{ // reset
		Access:  MethodAccessAllowed,
		Handler: pg.reset,
	}
	methods[2] = MethodDescriptor{ // capture
		Access:     MethodAccessAllowed,
		ParamTypes: []reflect.Type{captureParamType},
		Handler:    pg.capture,
	}

	return pg, nil
}

func (pg *ProfileGeneric) reset(_ []interface{}) (interface{}, error) {
	bufferAttr, ok := pg.Attributes[2]
	if !ok {
		return nil, ErrAttributeNotSupported
	}

	if bufferAttr.Type.Kind() == reflect.Slice {
		bufferAttr.Value = reflect.MakeSlice(bufferAttr.Type, 0, 0).Interface()
	} else {
		bufferAttr.Value = reflect.Zero(bufferAttr.Type).Interface()
	}
	pg.Attributes[2] = bufferAttr

	entriesAttr, ok := pg.Attributes[7]
	if !ok {
		return nil, ErrAttributeNotSupported
	}
	entriesAttr.Value = uint32(0)
	pg.Attributes[7] = entriesAttr

	return nil, nil
}

func (pg *ProfileGeneric) capture(params []interface{}) (interface{}, error) {
	bufferAttr, ok := pg.Attributes[2]
	if !ok {
		return nil, ErrAttributeNotSupported
	}
	if bufferAttr.Type.Kind() != reflect.Slice {
		return nil, ErrInvalidValueType
	}

	bufferVal := reflect.ValueOf(bufferAttr.Value)
	if !bufferVal.IsValid() || bufferVal.Kind() != reflect.Slice || bufferVal.IsNil() {
		bufferVal = reflect.MakeSlice(bufferAttr.Type, 0, 0)
	}

	newElem := reflect.ValueOf(params[0])
	if !newElem.Type().AssignableTo(bufferAttr.Type.Elem()) {
		return nil, ErrInvalidParameter
	}

	bufferVal = reflect.Append(bufferVal, newElem)
	bufferAttr.Value = bufferVal.Interface()
	pg.Attributes[2] = bufferAttr

	entriesAttr, ok := pg.Attributes[7]
	if !ok {
		return nil, ErrAttributeNotSupported
	}
	entriesAttr.Value = uint32(bufferVal.Len())
	pg.Attributes[7] = entriesAttr

	return nil, nil
}
