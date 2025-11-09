package cosem

import "reflect"

// AssociationLNClassID is the class ID for the "Association LN" interface class.
const AssociationLNClassID uint16 = 15

// AssociationLNVersion is the version of the "Association LN" interface class.
const AssociationLNVersion byte = 0

// AssociationLN represents the COSEM "Association LN" interface class.
type AssociationLN struct {
	BaseImpl
}

// ObjectListElement represents an element in the object_list attribute of the Association LN class.
type ObjectListElement struct {
	ClassID      uint16
	Version      uint8
	InstanceID   ObisCode
	AccessRights AccessRights
}

// AccessRights represents the access_rights attribute of an ObjectListElement.
type AccessRights struct {
	AttributeAccess []AttributeAccess
	MethodAccess    []MethodAccess
}

// NewAssociationLN creates a new instance of the "Association LN" interface class.
func NewAssociationLN(obis ObisCode) (*AssociationLN, error) {
	attributes := map[byte]AttributeDescriptor{
		1: { // logical_name
			Type:   reflect.TypeOf(ObisCode{}),
			Access: AttributeRead,
			Value:  obis,
		},
		2: { // object_list
			Type:   reflect.TypeOf([]ObjectListElement{}),
			Access: AttributeRead,
			Value:  []ObjectListElement{},
		},
	}

	return &AssociationLN{
		BaseImpl: BaseImpl{
			ClassID:    AssociationLNClassID,
			InstanceID: obis,
			Attributes: attributes,
			Methods:    map[byte]MethodDescriptor{},
		},
	}, nil
}

// AddObject adds a new object to the object_list attribute.
func (a *AssociationLN) AddObject(obj BaseInterface) {
	objList := a.Attributes[2].Value.([]ObjectListElement)

	// Get attribute and method access rights
	attrAccess := []AttributeAccess{}
	for i := byte(1); i <= 20; i++ { // Assuming max 20 attributes
		attrAccess = append(attrAccess, obj.GetAttributeAccess(i))
	}
	methodAccess := []MethodAccess{}
	for i := byte(1); i <= 20; i++ { // Assuming max 20 methods
		methodAccess = append(methodAccess, obj.GetMethodAccess(i))
	}

	objList = append(objList, ObjectListElement{
		ClassID:    obj.GetClassID(),
		Version:    0, // Version is not available in BaseInterface, default to 0
		InstanceID: obj.GetInstanceID(),
		AccessRights: AccessRights{
			AttributeAccess: attrAccess,
			MethodAccess:    methodAccess,
		},
	})

	a.Attributes[2] = AttributeDescriptor{
		Type:   reflect.TypeOf([]ObjectListElement{}),
		Access: AttributeRead,
		Value:  objList,
	}
}
