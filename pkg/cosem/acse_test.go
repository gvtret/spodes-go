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

func TestACSE_HandleAARQ(t *testing.T) {
	acse := NewACSE("password")

	t.Run("Successful Association", func(t *testing.T) {
		authValue, _ := asn1.Marshal(AuthenticationValue{GraphicString: "password"})
		aarq := &AARQ{
			ApplicationContextName: OidApplicationContextLN,
			MechanismName:          OidMechanismLLS,
			CallingAuthenticationValue: asn1.RawValue{
				Bytes: authValue,
			},
		}

		aare := acse.HandleAARQ(aarq)
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

		aare := acse.HandleAARQ(aarq)
		assert.Equal(t, ResultRejectedPermanent, aare.Result)
		assert.Equal(t, ACSEUserAuthenticationFailed, aare.ResultSourceDiagnostic.ACSEServiceUser)
	})
}

func TestACSE_HandleRLRQ(t *testing.T) {
	acse := NewACSE("password")
	acse.state = StateAssociated

	rlrq := &RLRQ{
		Reason: ReasonNormal,
	}

	rlre := acse.HandleRLRQ(rlrq)
	assert.Equal(t, ReasonNormal, rlre.Reason)
	assert.Equal(t, StateUnassociated, acse.state)
}
