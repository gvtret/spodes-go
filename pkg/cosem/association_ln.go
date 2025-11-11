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
	AttributeAccess []AttributeAccessItem
	MethodAccess    []MethodAccessItem
}

// AttributeAccessItem defines the access rights for a single attribute.
type AttributeAccessItem struct {
	AttributeID int
	AccessMode  AttributeAccessRight
}

// MethodAccessItem defines the access rights for a single method.
type MethodAccessItem struct {
	MethodID   int
	AccessMode MethodAccessRight
}

// AttributeAccessRight defines the possible access rights for an attribute.
type AttributeAccessRight byte

const (
	NoAccess  AttributeAccessRight = 0
	Read      AttributeAccessRight = 1
	Write     AttributeAccessRight = 2
	ReadWrite AttributeAccessRight = 3
)

// MethodAccessRight defines the possible access rights for a method.
type MethodAccessRight byte

const (
	Access    MethodAccessRight = 1
	NoAccess_ MethodAccessRight = 0
)

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
	attrAccessItems := []AttributeAccessItem{}
	for i := byte(1); i <= 20; i++ { // Assuming max 20 attributes
		access := obj.GetAttributeAccess(i)
		if access != AttributeNoAccess {
			var accessMode AttributeAccessRight
			if access&AttributeRead != 0 && access&AttributeWrite != 0 {
				accessMode = ReadWrite
			} else if access&AttributeRead != 0 {
				accessMode = Read
			} else if access&AttributeWrite != 0 {
				accessMode = Write
			}
			attrAccessItems = append(attrAccessItems, AttributeAccessItem{
				AttributeID: int(i),
				AccessMode:  accessMode,
			})
		}
	}

	methodAccessItems := []MethodAccessItem{}
	for i := byte(1); i <= 20; i++ { // Assuming max 20 methods
		access := obj.GetMethodAccess(i)
		if access != MethodNoAccess {
			methodAccessItems = append(methodAccessItems, MethodAccessItem{
				MethodID:   int(i),
				AccessMode: Access,
			})
		}
	}

	objList = append(objList, ObjectListElement{
		ClassID:    obj.GetClassID(),
		Version:    0, // Version is not available in BaseInterface, default to 0
		InstanceID: obj.GetInstanceID(),
		AccessRights: AccessRights{
			AttributeAccess: attrAccessItems,
			MethodAccess:    methodAccessItems,
		},
	})

	a.Attributes[2] = AttributeDescriptor{
		Type:   reflect.TypeOf([]ObjectListElement{}),
		Access: AttributeRead,
		Value:  objList,
	}
}

// CheckAttributeAccess verifies if a specific attribute has the required access rights.
func (a *AssociationLN) CheckAttributeAccess(obis ObisCode, attributeID byte, requiredAccess AttributeAccessRight) bool {
	objListAttr, ok := a.Attributes[2]
	if !ok {
		return false
	}
	objList, ok := objListAttr.Value.([]ObjectListElement)
	if !ok {
		return false
	}

	for _, elem := range objList {
		if elem.InstanceID.String() == obis.String() {
			for _, attrAccess := range elem.AccessRights.AttributeAccess {
				if attrAccess.AttributeID == int(attributeID) {
					if requiredAccess == Read {
						return attrAccess.AccessMode == Read || attrAccess.AccessMode == ReadWrite
					}
					if requiredAccess == Write {
						return attrAccess.AccessMode == Write || attrAccess.AccessMode == ReadWrite
					}
					return false
				}
			}
		}
	}
	return false // If the object or attribute is not in the list, access is denied.
}

// CheckMethodAccess verifies if a specific method is accessible.
func (a *AssociationLN) CheckMethodAccess(obis ObisCode, methodID byte) bool {
	objListAttr, ok := a.Attributes[2]
	if !ok {
		return false
	}
	objList, ok := objListAttr.Value.([]ObjectListElement)
	if !ok {
		return false
	}

	for _, elem := range objList {
		if elem.InstanceID.String() == obis.String() {
			for _, methodAccess := range elem.AccessRights.MethodAccess {
				if methodAccess.MethodID == int(methodID) {
					return methodAccess.AccessMode == Access
				}
			}
		}
	}
	return false // If the object or method is not in the list, access is denied.
}
