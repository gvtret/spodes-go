package cosem

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
)

// OIDs for COSEM application contexts and authentication mechanisms.
var (
	// OidApplicationContextLN specifies the object identifier for a Logical Name (LN) referencing association.
	OidApplicationContextLN = asn1.ObjectIdentifier{2, 16, 756, 5, 8, 1, 1}
	// OidApplicationContextSN specifies the object identifier for a Short Name (SN) referencing association.
	OidApplicationContextSN = asn1.ObjectIdentifier{2, 16, 756, 5, 8, 1, 2}
	// OidMechanismLLS specifies the object identifier for Low-Level Security (LLS) authentication.
	OidMechanismLLS = asn1.ObjectIdentifier{2, 16, 756, 5, 8, 2, 1}
	// OidMechanismHLS specifies the object identifier for High-Level Security (HLS) GMAC authentication.
	OidMechanismHLS = asn1.ObjectIdentifier{2, 16, 756, 5, 8, 2, 5}
)

// AssociationState represents the state of the COSEM association.
type AssociationState int

const (
	StateUnassociated AssociationState = iota
	StateAssociated
)

// ACSE manages the state of a COSEM association.
type ACSE struct {
	state             AssociationState
	password          string
	privateKey        *ecdsa.PrivateKey
	serverSystemTitle []byte
}

// NewACSE creates a new ACSE manager.
func NewACSE(password string, privateKey *ecdsa.PrivateKey, serverSystemTitle []byte) *ACSE {
	return &ACSE{
		state:             StateUnassociated,
		password:          password,
		privateKey:        privateKey,
		serverSystemTitle: serverSystemTitle,
	}
}

// AARQ (Association Request) APDU structure, used to initiate a COSEM association.
// It is encoded using ASN.1 BER rules.
type AARQ struct {
	ProtocolVersion            asn1.BitString        `asn1:"tag:0,optional,default:0"`
	ApplicationContextName     asn1.ObjectIdentifier `asn1:"tag:1"`
	CallingAPtitle             asn1.RawValue         `asn1:"tag:2,optional"`
	RespondingAPtitle          asn1.RawValue         `asn1:"tag:3,optional"`
	MechanismName              asn1.ObjectIdentifier `asn1:"tag:11,optional"`
	CallingAuthenticationValue asn1.RawValue         `asn1:"tag:12,explicit"` // CHOICE of Authentication-value
	UserInformation            asn1.RawValue         `asn1:"tag:30,optional"`
}

// AARE (Association Response) APDU structure, sent in response to an AARQ.
// It is encoded using ASN.1 BER rules.
type AARE struct {
	ProtocolVersion               asn1.BitString         `asn1:"tag:0,optional,default:0"`
	ApplicationContextName        asn1.ObjectIdentifier  `asn1:"tag:1"`
	Result                        asn1.Enumerated        `asn1:"tag:2"`
	ResultSourceDiagnostic        ResultSourceDiagnostic `asn1:"tag:3,explicit"`
	RespondingAPtitle             asn1.RawValue          `asn1:"tag:5,optional"`
	MechanismName                 asn1.ObjectIdentifier  `asn1:"tag:9,optional"`
	RespondingAuthenticationValue asn1.RawValue          `asn1:"tag:10,explicit,optional"` // CHOICE of Authentication-value
	UserInformation               asn1.RawValue          `asn1:"tag:30,optional"`
}

// RLRQ (Release Request) APDU structure, used to terminate a COSEM association.
type RLRQ struct {
	Reason          asn1.Enumerated `asn1:"tag:0,optional"`
	UserInformation asn1.RawValue   `asn1:"tag:30,optional"`
}

// RLRE (Release Response) APDU structure, sent in response to an RLRQ.
type RLRE struct {
	Reason          asn1.Enumerated `asn1:"tag:0,optional"`
	UserInformation asn1.RawValue   `asn1:"tag:30,optional"`
}

// AuthenticationValue represents the CHOICE for authentication data.
// For LLS, the GraphicString variant is used.
type AuthenticationValue struct {
	GraphicString string `asn1:"tag:0"`
}

// HLSAuthentication represents the authentication value for HLS.
type HLSAuthentication struct {
	EphemeralPublicKey []byte `asn1:"tag:0"`
}

// ResultSourceDiagnostic represents the CHOICE for association result diagnostics.
type ResultSourceDiagnostic struct {
	ACSEServiceUser     asn1.Enumerated `asn1:"tag:1,optional"`
	ACSEServiceProvider asn1.Enumerated `asn1:"tag:2,optional"`
}

// InitiateRequest represents the user-information field of an AARQ APDU.
type InitiateRequest struct {
	ProposedConformance       asn1.BitString `asn1:"optional"`
	ProposedMaxPduSize        int            `asn1:"optional"`
	ProposedDlmsVersionNumber int            `asn1:"optional"`
}

// InitiateResponse represents the user-information field of an AARE APDU.
type InitiateResponse struct {
	NegotiatedConformance       asn1.BitString  `asn1:"optional"`
	NegotiatedMaxPduSize        int             `asn1:"optional"`
	NegotiatedDlmsVersionNumber int             `asn1:"optional"`
	VAAname                     asn1.Enumerated `asn1:"optional"`
}

// Enumerated values for AARE.Result field.
const (
	ResultAccepted          asn1.Enumerated = 0
	ResultRejectedPermanent asn1.Enumerated = 1
	ResultRejectedTransient asn1.Enumerated = 2
)

// Enumerated values for ResultSourceDiagnostic.ACSEServiceUser.
const (
	ACSEUserNull                                asn1.Enumerated = 0
	ACSEUserNoReasonGiven                       asn1.Enumerated = 1
	ACSEUserAppContextNotSupported              asn1.Enumerated = 2
	ACSEUserAuthenticationFailed                asn1.Enumerated = 14
	ACSEUserAuthenticationMechanismNotSupported asn1.Enumerated = 15
)

// Enumerated values for ResultSourceDiagnostic.ACSEServiceProvider.
const (
	ACSEServiceProviderNull                asn1.Enumerated = 0
	ACSEServiceProviderNoReasonGiven       asn1.Enumerated = 1
	ACSEServiceProviderNoCommonACSEVersion asn1.Enumerated = 2
)

// Enumerated values for RLRQ.Reason and RLRE.Reason fields.
const (
	ReasonNormal      asn1.Enumerated = 0
	ReasonUrgent      asn1.Enumerated = 1
	ReasonUserDefined asn1.Enumerated = 30
)

// HandleAARQ processes an AARQ and returns an AARE.
func (a *ACSE) HandleAARQ(req *AARQ, securitySetup *SecuritySetup) (*AARE, error) {
	resp := &AARE{
		ApplicationContextName: req.ApplicationContextName,
		MechanismName:          req.MechanismName,
	}

	if !req.ApplicationContextName.Equal(OidApplicationContextLN) && !req.ApplicationContextName.Equal(OidApplicationContextSN) {
		resp.Result = ResultRejectedPermanent
		resp.ResultSourceDiagnostic = ResultSourceDiagnostic{
			ACSEServiceUser: ACSEUserAppContextNotSupported,
		}
		return resp, nil
	}

	if req.MechanismName.Equal(OidMechanismLLS) {
		var authVal AuthenticationValue
		_, err := asn1.Unmarshal(req.CallingAuthenticationValue.Bytes, &authVal)
		if err != nil {
			resp.Result = ResultRejectedPermanent
			resp.ResultSourceDiagnostic = ResultSourceDiagnostic{
				ACSEServiceUser: ACSEUserAuthenticationFailed,
			}
			return resp, nil
		}

		if authVal.GraphicString != a.password {
			resp.Result = ResultRejectedPermanent
			resp.ResultSourceDiagnostic = ResultSourceDiagnostic{
				ACSEServiceUser: ACSEUserAuthenticationFailed,
			}
			return resp, nil
		}
	} else if req.MechanismName.Equal(OidMechanismHLS) {
		if a.privateKey == nil {
			resp.Result = ResultRejectedPermanent
			resp.ResultSourceDiagnostic = ResultSourceDiagnostic{
				ACSEServiceUser: ACSEUserAuthenticationMechanismNotSupported,
			}
			return resp, nil
		}

		var authVal HLSAuthentication
		_, err := asn1.Unmarshal(req.CallingAuthenticationValue.Bytes, &authVal)
		if err != nil {
			resp.Result = ResultRejectedPermanent
			resp.ResultSourceDiagnostic = ResultSourceDiagnostic{
				ACSEServiceUser: ACSEUserAuthenticationFailed,
			}
			return resp, nil
		}
		// In a real implementation, the public key would be parsed and used for ECDH
		// For now, we'll just accept any public key.

		// Key agreement
		clientEphemeralPublicKey, err := UnmarshalPublicKey(authVal.EphemeralPublicKey)
		if err != nil {
			return nil, err
		}
		sharedSecret, err := ECDH(a.privateKey, clientEphemeralPublicKey)
		if err != nil {
			return nil, err
		}

		// Key derivation
		guek, gak, err := deriveKeys(sharedSecret)
		if err != nil {
			return nil, err
		}
		securitySetup.GlobalUnicastKey = guek
		securitySetup.GlobalAuthenticationKey = gak

	} else {
		resp.Result = ResultRejectedPermanent
		resp.ResultSourceDiagnostic = ResultSourceDiagnostic{
			ACSEServiceUser: ACSEUserAuthenticationMechanismNotSupported,
		}
		return resp, nil
	}

	a.state = StateAssociated
	resp.Result = ResultAccepted
	resp.ResultSourceDiagnostic = ResultSourceDiagnostic{
		ACSEServiceUser: ACSEUserNull,
	}
	return resp, nil
}

// HandleRLRQ processes an RLRQ and returns an RLRE.
func (a *ACSE) HandleRLRQ(req *RLRQ) *RLRE {
	a.state = StateUnassociated
	return &RLRE{
		Reason: req.Reason,
	}
}

// deriveKeys derives the GUEK and GAK from the shared secret.
func deriveKeys(sharedSecret []byte) ([]byte, []byte, error) {
	// For simplicity, we'll use a simple key derivation function.
	// In a real implementation, a proper KDF like HKDF should be used.
	hash := sha256.Sum256(sharedSecret)
	guek := hash[:16]
	gak := hash[16:]
	return guek, gak, nil
}

// Encode encodes the AARQ APDU into a byte slice.
func (a *AARQ) Encode() ([]byte, error) {
	return asn1.Marshal(*a)
}

// Decode decodes a byte slice into an AARQ APDU.
func (a *AARQ) Decode(src []byte) error {
	_, err := asn1.Unmarshal(src, a)
	return err
}

// Encode encodes the AARE APDU into a byte slice.
func (a *AARE) Encode() ([]byte, error) {
	return asn1.Marshal(*a)
}

// Decode decodes a byte slice into an AARE APDU.
func (a *AARE) Decode(src []byte) error {
	_, err := asn1.Unmarshal(src, a)
	return err
}

// Encode encodes the RLRQ APDU into a byte slice.
func (r *RLRQ) Encode() ([]byte, error) {
	return asn1.Marshal(*r)
}

// Decode decodes a byte slice into an RLRQ APDU.
func (r *RLRQ) Decode(src []byte) error {
	_, err := asn1.Unmarshal(src, r)
	return err
}

// Encode encodes the RLRE APDU into a byte slice.
func (r *RLRE) Encode() ([]byte, error) {
	return asn1.Marshal(*r)
}

// Decode decodes a byte slice into an RLRE APDU.
func (r *RLRE) Decode(src []byte) error {
	_, err := asn1.Unmarshal(src, r)
	return err
}
