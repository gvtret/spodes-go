package cosem

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
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
	assert.Len(t, sig, 64)

	err = VerifyECDSA(pub, msg, sig)
	assert.NoError(t, err)
}

func TestECDSASignaturePadding(t *testing.T) {
	curve := elliptic.P256()
	d := big.NewInt(1)
	x, y := curve.ScalarBaseMult(d.Bytes())
	priv := &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: curve, X: x, Y: y}, D: d}
	msg := make([]byte, 32)
	msg[31] = 1

	rHex := "46c5304f0e9310c0dee40831057c537d7151500940a4ca1251c00facd1e0c7db"
	sHex := "00363e3a7cfa4eff0b65db3bee633f19ab42ed13f4b250514ce4373b890f3c62"

	r := new(big.Int)
	_, ok := r.SetString(rHex, 16)
	assert.True(t, ok)
	s := new(big.Int)
	_, ok = s.SetString(sHex, 16)
	assert.True(t, ok)

	coordinateSize := (curve.Params().BitSize + 7) / 8
	rBytes, err := padScalar(r.Bytes(), coordinateSize)
	assert.NoError(t, err)
	sBytes, err := padScalar(s.Bytes(), coordinateSize)
	assert.NoError(t, err)

	sig := append(append(make([]byte, 0, 2*coordinateSize), rBytes...), sBytes...)
	assert.Equal(t, 64, len(sig))
	assert.Equal(t, byte(0x00), sig[32])

	err = VerifyECDSA(&priv.PublicKey, msg, sig)
	assert.NoError(t, err)

	err = VerifyECDSA(&priv.PublicKey, msg, sig[:len(sig)-1])
	assert.ErrorIs(t, err, ErrInvalidSignature)
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
