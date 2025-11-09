package cosem

import (
	"encoding/asn1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAARQ_EncodeDecode(t *testing.T) {
	authValue, _ := asn1.Marshal(AuthenticationValue{GraphicString: "password"})
	aarq := &AARQ{
		ApplicationContextName: OidApplicationContextLN,
		MechanismName:          OidMechanismLLS,
		CallingAuthenticationValue: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        12,
			IsCompound: true,
			Bytes:      authValue,
		},
	}

	encoded, err := aarq.Encode()
	assert.NoError(t, err)

	decoded := &AARQ{}
	err = decoded.Decode(encoded)
	assert.NoError(t, err)

	assert.True(t, aarq.ApplicationContextName.Equal(decoded.ApplicationContextName))
	assert.True(t, aarq.MechanismName.Equal(decoded.MechanismName))
	assert.Equal(t, aarq.CallingAuthenticationValue.Bytes, decoded.CallingAuthenticationValue.Bytes)
}

func TestAARE_EncodeDecode(t *testing.T) {
	aare := &AARE{
		ApplicationContextName: OidApplicationContextLN,
		Result:                 ResultAccepted,
		ResultSourceDiagnostic: ResultSourceDiagnostic{
			ACSEServiceUser: ACSEUserNull,
		},
	}

	encoded, err := aare.Encode()
	assert.NoError(t, err)

	decoded := &AARE{}
	err = decoded.Decode(encoded)
	assert.NoError(t, err)

	assert.True(t, aare.ApplicationContextName.Equal(decoded.ApplicationContextName))
	assert.Equal(t, aare.Result, decoded.Result)
	assert.Equal(t, aare.ResultSourceDiagnostic.ACSEServiceUser, decoded.ResultSourceDiagnostic.ACSEServiceUser)
}

func TestRLRQ_EncodeDecode(t *testing.T) {
	rlrq := &RLRQ{
		Reason: ReasonNormal,
	}

	encoded, err := rlrq.Encode()
	assert.NoError(t, err)

	decoded := &RLRQ{}
	err = decoded.Decode(encoded)
	assert.NoError(t, err)

	assert.Equal(t, rlrq.Reason, decoded.Reason)
}

func TestRLRE_EncodeDecode(t *testing.T) {
	rlre := &RLRE{
		Reason: ReasonNormal,
	}

	encoded, err := rlre.Encode()
	assert.NoError(t, err)

	decoded := &RLRE{}
	err = decoded.Decode(encoded)
	assert.NoError(t, err)

	assert.Equal(t, rlre.Reason, decoded.Reason)
}

func TestACSE_HandleAARQ_LLS(t *testing.T) {
	acse := NewACSE("password", nil, nil)

	t.Run("Successful Association", func(t *testing.T) {
		authValue, _ := asn1.Marshal(AuthenticationValue{GraphicString: "password"})
		aarq := &AARQ{
			ApplicationContextName: OidApplicationContextLN,
			MechanismName:          OidMechanismLLS,
			CallingAuthenticationValue: asn1.RawValue{
				Bytes: authValue,
			},
		}

		aare, err := acse.HandleAARQ(aarq, nil)
		assert.NoError(t, err)
		assert.Equal(t, ResultAccepted, aare.Result)
		assert.Equal(t, StateAssociated, acse.state)
	})

	t.Run("Failed Authentication", func(t *testing.T) {
		authValue, _ := asn1.Marshal(AuthenticationValue{GraphicString: "wrong_password"})
		aarq := &AARQ{
			ApplicationContextName: OidApplicationContextLN,
			MechanismName:          OidMechanismLLS,
			CallingAuthenticationValue: asn1.RawValue{
				Bytes: authValue,
			},
		}

		aare, err := acse.HandleAARQ(aarq, nil)
		assert.NoError(t, err)
		assert.Equal(t, ResultRejectedPermanent, aare.Result)
		assert.Equal(t, ACSEUserAuthenticationFailed, aare.ResultSourceDiagnostic.ACSEServiceUser)
	})
}

func TestACSE_HandleAARQ_HLS(t *testing.T) {
	priv, pub, err := GenerateECDHKeys()
	assert.NoError(t, err)
	acse := NewACSE("", priv, []byte("SERVER01"))

	clientPriv, clientPub, err := GenerateECDHKeys()
	assert.NoError(t, err)
	marshaledClientPub, _ := MarshalPublicKey(clientPub)
	authValue, _ := asn1.Marshal(HLSAuthentication{EphemeralPublicKey: marshaledClientPub})
	aarq := &AARQ{
		ApplicationContextName: OidApplicationContextLN,
		MechanismName:          OidMechanismHLS,
		CallingAuthenticationValue: asn1.RawValue{
			Bytes: authValue,
		},
	}
	obis, _ := NewObisCodeFromString("0.0.43.0.0.255")
	securitySetup, _ := NewSecuritySetup(*obis, []byte("CLIENT"), []byte("SERVER01"), nil, nil, nil)

	aare, err := acse.HandleAARQ(aarq, securitySetup)
	assert.NoError(t, err)
	assert.Equal(t, ResultAccepted, aare.Result)

	sharedSecret, _ := ECDH(clientPriv, pub)
	guek, gak, _ := deriveKeys(sharedSecret)
	assert.Equal(t, guek, securitySetup.GlobalUnicastKey)
	assert.Equal(t, gak, securitySetup.GlobalAuthenticationKey)
}

func TestACSE_HandleRLRQ(t *testing.T) {
	acse := NewACSE("password", nil, nil)
	acse.state = StateAssociated

	rlrq := &RLRQ{
		Reason: ReasonNormal,
	}

	rlre := acse.HandleRLRQ(rlrq)
	assert.Equal(t, ReasonNormal, rlre.Reason)
	assert.Equal(t, StateUnassociated, acse.state)
}
