package cosem

import (
	"reflect"
)

// SecuritySetupClassID is the class ID for the "Security setup" interface class.
const SecuritySetupClassID uint16 = 64

// SecuritySetupVersion is the version of the "Security setup" interface class.
const SecuritySetupVersion byte = 0

// SecurityPolicy represents the security_policy attribute of the Security setup class.
// It's a bitmask defining the minimum security level for requests and responses.
type SecurityPolicy byte

const (
	PolicyNone                    SecurityPolicy = 0x00
	PolicyAuthenticatedRequest    SecurityPolicy = 0x04 // bit 2
	PolicyEncryptedRequest        SecurityPolicy = 0x08 // bit 3
	PolicyDigitallySignedRequest  SecurityPolicy = 0x10 // bit 4
	PolicyAuthenticatedResponse   SecurityPolicy = 0x20 // bit 5
	PolicyEncryptedResponse       SecurityPolicy = 0x40 // bit 6
	PolicyDigitallySignedResponse SecurityPolicy = 0x80 // bit 7
)

// SecuritySuite represents the security_suite attribute of the Security setup class.
type SecuritySuite byte

const (
	SecuritySuite0 SecuritySuite = 0 // AES-GCM-128
	SecuritySuite1 SecuritySuite = 1 // AES-128-CBC with GMAC
	SecuritySuite2 SecuritySuite = 2 // AES-256-CBC with GMAC
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
			Type:   reflect.TypeOf(PolicyNone),
			Access: AttributeRead | AttributeWrite,
			Value:  PolicyNone,
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
