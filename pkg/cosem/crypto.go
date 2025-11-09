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
	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, priv.D.Bytes())
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
	return append(r.Bytes(), s.Bytes()...), nil
}

// VerifyECDSA verifies a signature using the provided public key.
func VerifyECDSA(pub *ecdsa.PublicKey, msg, sig []byte) error {
	if pub == nil {
		return ErrInvalidPublicKey
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	if !ecdsa.Verify(pub, msg, r, s) {
		return ErrInvalidSignature
	}
	return nil
}

// MarshalPublicKey marshals a public key into a byte slice.
func MarshalPublicKey(pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil {
		return nil, ErrInvalidPublicKey
	}
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y), nil
}

// UnmarshalPublicKey unmarshals a byte slice into a public key.
func UnmarshalPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(elliptic.P256(), data)
	if x == nil {
		return nil, ErrInvalidPublicKey
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}
