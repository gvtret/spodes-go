package cosem

import (
	"reflect"
)

// SecuritySetupClassID is the class ID for the "Security setup" interface class.
const SecuritySetupClassID uint16 = 64

// SecuritySetupVersion is the version of the "Security setup" interface class.
const SecuritySetupVersion byte = 0

// SecurityPolicy represents the security_policy attribute of the Security setup class.
type SecurityPolicy byte

const (
	SecurityPolicyNone                      SecurityPolicy = 0
	SecurityPolicyAuthenticated             SecurityPolicy = 1
	SecurityPolicyEncrypted                 SecurityPolicy = 2
	SecurityPolicyAuthenticatedAndEncrypted SecurityPolicy = 3
)

// SecuritySuite represents the security_suite attribute of the Security setup class.
type SecuritySuite byte

const (
	SecuritySuite0 SecuritySuite = 0
)

// SecuritySetup represents the COSEM "Security setup" interface class.
type SecuritySetup struct {
	BaseImpl
	MasterKey               []byte // KEK
	GlobalUnicastKey        []byte // GUEK
	GlobalAuthenticationKey []byte // GAK
}

// NewSecuritySetup creates a new instance of the "Security setup" interface class.
func NewSecuritySetup(obis ObisCode, clientSystemTitle []byte, serverSystemTitle []byte, masterKey, guek, gak []byte) (*SecuritySetup, error) {
	attributes := map[byte]AttributeDescriptor{
		1: { // logical_name
			Type:   reflect.TypeOf(ObisCode{}),
			Access: AttributeRead,
			Value:  obis,
		},
		2: { // security_policy
			Type:   reflect.TypeOf(SecurityPolicyNone),
			Access: AttributeRead | AttributeWrite,
			Value:  SecurityPolicyNone,
		},
		3: { // security_suite
			Type:   reflect.TypeOf(SecuritySuite0),
			Access: AttributeRead | AttributeWrite,
			Value:  SecuritySuite0,
		},
		4: { // client_system_title
			Type:   reflect.TypeOf(clientSystemTitle),
			Access: AttributeRead,
			Value:  clientSystemTitle,
		},
		5: { // server_system_title
			Type:   reflect.TypeOf(serverSystemTitle),
			Access: AttributeRead,
			Value:  serverSystemTitle,
		},
	}

	return &SecuritySetup{
		BaseImpl: BaseImpl{
			ClassID:    SecuritySetupClassID,
			InstanceID: obis,
			Attributes: attributes,
			Methods:    map[byte]MethodDescriptor{},
		},
		MasterKey:               masterKey,
		GlobalUnicastKey:        guek,
		GlobalAuthenticationKey: gak,
	}, nil
}
