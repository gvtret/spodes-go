package cosem

import (
	"encoding/hex"
	"github.com/aead/cmac"
	"github.com/ddulesov/gogost/gost3412128"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKuznyechikEncrypt(t *testing.T) {
	key, _ := hex.DecodeString("8899aabbccddeeff0011223344556677fedcba98765432100123456789abcdef")
	plaintext, _ := hex.DecodeString("1122334455667700ffeeddccbbaa9988")
	expected, _ := hex.DecodeString("7f679d90bebc24305a468d42b9d4edcd")

	c := gost3412128.NewCipher(key)

	ciphertext := make([]byte, 16)
	c.Encrypt(ciphertext, plaintext)

	assert.Equal(t, expected, ciphertext)
}

func TestKuznyechikDecrypt(t *testing.T) {
	key, _ := hex.DecodeString("8899aabbccddeeff0011223344556677fedcba98765432100123456789abcdef")
	ciphertext, _ := hex.DecodeString("7f679d90bebc24305a468d42b9d4edcd")
	expected, _ := hex.DecodeString("1122334455667700ffeeddccbbaa9988")

	c := gost3412128.NewCipher(key)
	plaintext := make([]byte, 16)
	c.Decrypt(plaintext, ciphertext)

	assert.Equal(t, expected, plaintext)
}

func TestCMAC(t *testing.T) {
	// Test vector from GOST R 34.13-2015
	key, err := hex.DecodeString("8899aabbccddeeff0011223344556677fedcba98765432100123456789abcdef")
	assert.NoError(t, err)
	msg, err := hex.DecodeString("1122334455667700ffeeddccbbaa9988")
	assert.NoError(t, err)
	expected, err := hex.DecodeString("51aa8ebefe937200c21e2518bd4a2edb")
	assert.NoError(t, err)

	c := gost3412128.NewCipher(key)

	tag, err := cmac.Sum(msg, c, 16)
	assert.NoError(t, err)
	assert.Equal(t, expected, tag)
}

func TestCTREncrypt(t *testing.T) {
	key, _ := hex.DecodeString("8899aabbccddeeff0011223344556677fedcba98765432100123456789abcdef")
	iv, _ := hex.DecodeString("1122334455667700ffeeddccbbaa9988")
	plaintext, _ := hex.DecodeString("00000000000000000000000000000000")
	expected, _ := hex.DecodeString("7f679d90bebc24305a468d42b9d4edcd")

	ciphertext, err := ctrEncrypt(key, iv, plaintext)
	assert.NoError(t, err)
	assert.Equal(t, expected, ciphertext)
}

func TestEncryptAndTag_DecryptAndVerify_Suite3(t *testing.T) {
	key, _ := hex.DecodeString("0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF")
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")
	header := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext, err := EncryptAndTag(key, plaintext, serverSystemTitle, header, SecuritySuite3)
	assert.NoError(t, err)

	decrypted, err := DecryptAndVerify(key, ciphertext, serverSystemTitle, header, SecuritySuite3, 0)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
