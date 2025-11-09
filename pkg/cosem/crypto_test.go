package cosem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestECDH(t *testing.T) {
	privA, pubA, err := GenerateECDHKeys()
	assert.NoError(t, err)

	privB, pubB, err := GenerateECDHKeys()
	assert.NoError(t, err)

	secretA, err := ECDH(privA, pubB)
	assert.NoError(t, err)

	secretB, err := ECDH(privB, pubA)
	assert.NoError(t, err)

	assert.Equal(t, secretA, secretB)
}

func TestECDSA(t *testing.T) {
	priv, pub, err := GenerateECDHKeys()
	assert.NoError(t, err)

	msg := []byte("Hello, COSEM!")
	sig, err := SignECDSA(priv, msg)
	assert.NoError(t, err)

	err = VerifyECDSA(pub, msg, sig)
	assert.NoError(t, err)
}

func TestMarshalUnmarshalPublicKey(t *testing.T) {
	_, pub, err := GenerateECDHKeys()
	assert.NoError(t, err)

	marshaled, err := MarshalPublicKey(pub)
	assert.NoError(t, err)

	unmarshaled, err := UnmarshalPublicKey(marshaled)
	assert.NoError(t, err)

	assert.Equal(t, pub, unmarshaled)
}
