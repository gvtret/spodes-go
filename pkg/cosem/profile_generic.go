package cosem

import (
	"reflect"
)

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
			Type:   reflect.TypeOf(capturePeriod),
			Access: AttributeRead,
			Value:  capturePeriod,
		},
		5: { // sort_method
			Type:   reflect.TypeOf(sortMethod),
			Access: AttributeRead,
			Value:  sortMethod,
		},
		6: { // sort_object
			Type:   reflect.TypeOf(sortObject),
			Access: AttributeRead,
			Value:  sortObject,
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

	methods := map[byte]MethodDescriptor{}

	return &ProfileGeneric{
		BaseImpl: BaseImpl{
			ClassID:    ProfileGenericClassID,
			InstanceID: obis,
			Attributes: attributes,
			Methods:    methods,
		},
	}, nil
}
