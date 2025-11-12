package cosem

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
)

var (
	// ErrInvalidSignature is returned when a signature is invalid.
	ErrInvalidSignature = fmt.Errorf("invalid signature")
	// ErrKeyAgreementFailed is returned when key agreement fails.
	ErrKeyAgreementFailed = fmt.Errorf("key agreement failed")
	// ErrInvalidPublicKey is returned when a public key is invalid.
	ErrInvalidPublicKey = fmt.Errorf("invalid public key")
	// ErrInvalidPrivateKey is returned when a private key is invalid.
	ErrInvalidPrivateKey = fmt.Errorf("invalid private key")
)

// GenerateECDHKeys generates a new ECDH key pair.
func GenerateECDHKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv, &priv.PublicKey, nil
}

// ECDH performs a key agreement using the provided private and public keys.
func ECDH(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) ([]byte, error) {
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}
	if pub == nil {
		return nil, ErrInvalidPublicKey
	}
	x, _ := pub.ScalarMult(pub.X, pub.Y, priv.D.Bytes())
	if x == nil {
		return nil, ErrKeyAgreementFailed
	}
	return x.Bytes(), nil
}

// SignECDSA signs a message using the provided private key.
func SignECDSA(priv *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}
	r, s, err := ecdsa.Sign(rand.Reader, priv, msg)
	if err != nil {
		return nil, err
	}
	params := priv.Curve.Params()
	if params == nil {
		return nil, ErrInvalidPrivateKey
	}

	coordinateSize := (params.BitSize + 7) / 8

	rBytes, err := padScalar(r.Bytes(), coordinateSize)
	if err != nil {
		return nil, err
	}
	sBytes, err := padScalar(s.Bytes(), coordinateSize)
	if err != nil {
		return nil, err
	}

	sig := make([]byte, 0, 2*coordinateSize)
	sig = append(sig, rBytes...)
	sig = append(sig, sBytes...)
	return sig, nil
}

// VerifyECDSA verifies a signature using the provided public key.
func VerifyECDSA(pub *ecdsa.PublicKey, msg, sig []byte) error {
	if pub == nil {
		return ErrInvalidPublicKey
	}
	params := pub.Curve.Params()
	if params == nil {
		return ErrInvalidPublicKey
	}

	coordinateSize := (params.BitSize + 7) / 8
	if len(sig) != 2*coordinateSize {
		return ErrInvalidSignature
	}

	r := new(big.Int).SetBytes(sig[:coordinateSize])
	s := new(big.Int).SetBytes(sig[coordinateSize:])
	if !ecdsa.Verify(pub, msg, r, s) {
		return ErrInvalidSignature
	}
	return nil
}

func padScalar(b []byte, size int) ([]byte, error) {
	if len(b) > size {
		return nil, fmt.Errorf("scalar length %d exceeds size %d", len(b), size)
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out, nil
}

// MarshalPublicKey marshals a public key into a byte slice.
func MarshalPublicKey(pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil {
		return nil, ErrInvalidPublicKey
	}
	if pub.X == nil || pub.Y == nil || pub.Curve == nil {
		return nil, ErrInvalidPublicKey
	}

	params := pub.Params()
	if params == nil {
		return nil, ErrInvalidPublicKey
	}

	coordinateSize := (params.BitSize + 7) / 8
	encoded := make([]byte, 1+2*coordinateSize)
	encoded[0] = 0x04

	xBytes := pub.X.Bytes()
	yBytes := pub.Y.Bytes()

	if len(xBytes) > coordinateSize || len(yBytes) > coordinateSize {
		return nil, ErrInvalidPublicKey
	}

	copy(encoded[1+coordinateSize-len(xBytes):1+coordinateSize], xBytes)
	copy(encoded[1+2*coordinateSize-len(yBytes):], yBytes)

	return encoded, nil
}

// UnmarshalPublicKey unmarshals a byte slice into a public key.
func UnmarshalPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	curve := elliptic.P256()
	params := curve.Params()
	coordinateSize := (params.BitSize + 7) / 8
	expectedLen := 1 + 2*coordinateSize

	if len(data) != expectedLen || len(data) == 0 || data[0] != 0x04 {
		return nil, ErrInvalidPublicKey
	}

	x := new(big.Int).SetBytes(data[1 : 1+coordinateSize])
	y := new(big.Int).SetBytes(data[1+coordinateSize:])
	if !curve.IsOnCurve(x, y) {
		return nil, ErrInvalidPublicKey
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}
