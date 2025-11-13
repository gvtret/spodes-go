package cosem

import (
	"encoding/asn1"
	"reflect"
)

// AssociationLNClassID is the class ID for the "Association LN" interface class.
const AssociationLNClassID uint16 = 15

// AssociationLNVersion is the version of the "Association LN" interface class.
const AssociationLNVersion byte = 0

// AssociationLN represents the COSEM "Association LN" interface class.
type AssociationLN struct {
	BaseImpl
	serverInvocationCounter uint32
}

// ObjectListElement represents an element in the object_list attribute of the Association LN class.
type ObjectListElement struct {
	ClassID      uint16
	Version      uint8
	InstanceID   ObisCode
	AccessRights AccessRights
}

// AssociatedPartnersID represents the associated_partners_id attribute.
type AssociatedPartnersID struct {
	ClientSAP uint16
	ServerSAP uint16
}

// ApplicationContextName represents the application_context_name attribute.
type ApplicationContextName struct {
	ContextID   byte
	LogicalName ObisCode
}

// XDLMSContextInfo represents the xDLMS_context_info attribute.
type XDLMSContextInfo struct {
	Conformance       uint32
	MaxReceivePDUSize uint16
	MaxSendPDUSize    uint16
	DLMSVersionNumber byte
	QualityOfService  byte
	CypheringInfo     []byte
}

// AuthenticationMechanismName represents the authentication_mechanism_name attribute.
type AuthenticationMechanismName struct {
	MechanismID   byte
	MechanismName asn1.ObjectIdentifier
}

// AssociationStatus defines the association_status attribute values.
type AssociationStatus byte

const (
	// AssociationStatusNonAssociated indicates no active association.
	AssociationStatusNonAssociated AssociationStatus = iota
	// AssociationStatusAssociationPending indicates an association is in progress and awaiting HLS.
	AssociationStatusAssociationPending
	// AssociationStatusAssociated indicates an active association is established.
	AssociationStatusAssociated
)

// UserListEntry represents a single entry in the user_list attribute.
type UserListEntry struct {
	UserID   uint16
	UserRole byte
}

// AssociationInformation aggregates key attributes exposed by get_association_information.
type AssociationInformation struct {
	AssociatedPartnersID    AssociatedPartnersID
	ApplicationContextName  ApplicationContextName
	XDLMSContextInfo        XDLMSContextInfo
	AuthenticationMechanism AuthenticationMechanismName
	SecuritySetupReference  ObisCode
	ClientSAP               uint16
	ServerSAP               uint16
	UserList                []UserListEntry
}

const (
	associationLNMethodAssociate byte = iota + 1
	associationLNMethodReplyToHLS
	associationLNMethodGetAssociationInformation
	associationLNMethodGetApplicationContextNameList
)

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
		3: { // associated_partners_id
			Type:   reflect.TypeOf(AssociatedPartnersID{}),
			Access: AttributeRead,
			Value: AssociatedPartnersID{
				ClientSAP: 0,
				ServerSAP: 0,
			},
		},
		4: { // application_context_name
			Type:   reflect.TypeOf(ApplicationContextName{}),
			Access: AttributeRead,
			Value: ApplicationContextName{
				ContextID:   0,
				LogicalName: obis,
			},
		},
		5: { // xDLMS_context_info
			Type:   reflect.TypeOf(XDLMSContextInfo{}),
			Access: AttributeRead,
			Value: XDLMSContextInfo{
				Conformance:       0,
				MaxReceivePDUSize: 0,
				MaxSendPDUSize:    0,
				DLMSVersionNumber: 6,
				QualityOfService:  0,
				CypheringInfo:     []byte{},
			},
		},
		6: { // authentication_mechanism_name
			Type:   reflect.TypeOf(AuthenticationMechanismName{}),
			Access: AttributeRead | AttributeWrite,
			Value: AuthenticationMechanismName{
				MechanismID:   0,
				MechanismName: asn1.ObjectIdentifier{},
			},
		},
		7: { // secret
			Type:   reflect.TypeOf([]byte{}),
			Access: AttributeWrite,
			Value:  []byte{},
		},
		8: { // association_status
			Type:   reflect.TypeOf(AssociationStatus(0)),
			Access: AttributeRead,
			Value:  AssociationStatusNonAssociated,
		},
		9: { // security_setup_reference
			Type:   reflect.TypeOf(ObisCode{}),
			Access: AttributeRead,
			Value:  ObisCode{},
		},
		10: { // client_SAP
			Type:   reflect.TypeOf(uint16(0)),
			Access: AttributeRead,
			Value:  uint16(0),
		},
		11: { // server_SAP
			Type:   reflect.TypeOf(uint16(0)),
			Access: AttributeRead,
			Value:  uint16(0),
		},
		12: { // user_list
			Type:   reflect.TypeOf([]UserListEntry{}),
			Access: AttributeRead,
			Value:  []UserListEntry{},
		},
	}

	assoc := &AssociationLN{
		BaseImpl: BaseImpl{
			ClassID:    AssociationLNClassID,
			InstanceID: obis,
			Attributes: attributes,
			Methods:    map[byte]MethodDescriptor{},
		},
	}

	assoc.Methods[associationLNMethodAssociate] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: []reflect.Type{},
		ReturnType: reflect.TypeOf(AssociationStatus(0)),
		Handler:    assoc.handleAssociate,
	}

	assoc.Methods[associationLNMethodReplyToHLS] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: []reflect.Type{reflect.TypeOf([]byte{})},
		ReturnType: reflect.TypeOf(true),
		Handler:    assoc.handleReplyToHLSAuthentication,
	}

	assoc.Methods[associationLNMethodGetAssociationInformation] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: []reflect.Type{},
		ReturnType: reflect.TypeOf(AssociationInformation{}),
		Handler:    assoc.handleGetAssociationInformation,
	}

	assoc.Methods[associationLNMethodGetApplicationContextNameList] = MethodDescriptor{
		Access:     MethodAccessAllowed,
		ParamTypes: []reflect.Type{},
		ReturnType: reflect.TypeOf([]ApplicationContextName{}),
		Handler:    assoc.handleGetApplicationContextNameList,
	}

	return assoc, nil
}

func (a *AssociationLN) handleAssociate(_ []interface{}) (interface{}, error) {
	statusAttr := a.Attributes[8]
	status := statusAttr.Value.(AssociationStatus)
	if status != AssociationStatusNonAssociated {
		return nil, ErrAccessDenied
	}

	statusAttr.Value = AssociationStatusAssociationPending
	a.Attributes[8] = statusAttr
	return statusAttr.Value, nil
}

func (a *AssociationLN) handleReplyToHLSAuthentication(params []interface{}) (interface{}, error) {
	statusAttr := a.Attributes[8]
	status := statusAttr.Value.(AssociationStatus)
	if status != AssociationStatusAssociationPending {
		return nil, ErrAccessDenied
	}

	challenge := params[0].([]byte)
	if len(challenge) == 0 {
		return nil, ErrInvalidParameter
	}

	statusAttr.Value = AssociationStatusAssociated
	a.Attributes[8] = statusAttr

	return true, nil
}

func (a *AssociationLN) handleGetAssociationInformation(_ []interface{}) (interface{}, error) {
	statusAttr := a.Attributes[8]
	status := statusAttr.Value.(AssociationStatus)
	if status != AssociationStatusAssociated {
		return nil, ErrAccessDenied
	}

	userList := a.Attributes[12].Value.([]UserListEntry)
	copiedUserList := append([]UserListEntry{}, userList...)
	if copiedUserList == nil {
		copiedUserList = []UserListEntry{}
	}

	info := AssociationInformation{
		AssociatedPartnersID:    a.Attributes[3].Value.(AssociatedPartnersID),
		ApplicationContextName:  a.Attributes[4].Value.(ApplicationContextName),
		XDLMSContextInfo:        a.Attributes[5].Value.(XDLMSContextInfo),
		AuthenticationMechanism: a.Attributes[6].Value.(AuthenticationMechanismName),
		SecuritySetupReference:  a.Attributes[9].Value.(ObisCode),
		ClientSAP:               a.Attributes[10].Value.(uint16),
		ServerSAP:               a.Attributes[11].Value.(uint16),
		UserList:                copiedUserList,
	}

	return info, nil
}

func (a *AssociationLN) handleGetApplicationContextNameList(_ []interface{}) (interface{}, error) {
	ctx := a.Attributes[4].Value.(ApplicationContextName)
	return []ApplicationContextName{ctx}, nil
}

// SetServerInvocationCounter updates the last server-side invocation counter value.
// The next secured response from this association will use counter+1.
func (a *AssociationLN) SetServerInvocationCounter(counter uint32) {
	a.serverInvocationCounter = counter
}

// ServerInvocationCounter returns the last server-side invocation counter value.
func (a *AssociationLN) ServerInvocationCounter() uint32 {
	return a.serverInvocationCounter
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
