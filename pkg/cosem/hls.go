package cosem

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// APDUType constants for secured APDUs.
const (
	APDU_GLO_GET_REQUEST     APDUType = 0xC8
	APDU_GLO_SET_REQUEST     APDUType = 0xC9
	APDU_GLO_ACTION_REQUEST  APDUType = 0xCB
	APDU_GLO_GET_RESPONSE    APDUType = 0xCC
	APDU_GLO_SET_RESPONSE    APDUType = 0xCD
	APDU_GLO_ACTION_RESPONSE APDUType = 0xCF
)

// SecurityControl byte flags.
const (
	SecurityControlAuthenticationOnly        SecurityControl = 0x10
	SecurityControlEncryptionOnly            SecurityControl = 0x20
	SecurityControlAuthenticatedAndEncrypted SecurityControl = 0x30
)

// SecurityControl represents the security control byte.
type SecurityControl byte

// SecurityHeader represents the security header of a secured APDU.
type SecurityHeader struct {
	SecurityControl   SecurityControl
	FrameCounter      uint32
	AuthenticatedData []byte // Additional authenticated data
}

// ErrReplayAttack is returned when a replay attack is detected.
var ErrReplayAttack = fmt.Errorf("replay attack detected")

// ErrAuthenticationFailed is returned when authentication fails.
var ErrAuthenticationFailed = fmt.Errorf("authentication failed")

// ErrInvalidPadding is returned when PKCS#7 padding is invalid.
var ErrInvalidPadding = fmt.Errorf("invalid PKCS#7 padding")

// Encode encodes the SecurityHeader into a byte slice.
func (h *SecurityHeader) Encode() ([]byte, error) {
	buf := make([]byte, 5)
	buf[0] = byte(h.SecurityControl)
	buf[1] = byte(h.FrameCounter >> 24)
	buf[2] = byte(h.FrameCounter >> 16)
	buf[3] = byte(h.FrameCounter >> 8)
	buf[4] = byte(h.FrameCounter)
	return buf, nil
}

// Decode decodes a byte slice into a SecurityHeader.
func (h *SecurityHeader) Decode(src []byte) error {
	if len(src) < 5 {
		return fmt.Errorf("invalid security header length: got %d, want at least 5", len(src))
	}
	h.SecurityControl = SecurityControl(src[0])
	h.FrameCounter = uint32(src[1])<<24 | uint32(src[2])<<16 | uint32(src[3])<<8 | uint32(src[4])
	return nil
}

// EncryptAndTag encrypts and authenticates a plaintext APDU.
func EncryptAndTag(key, plaintext, serverSystemTitle []byte, header *SecurityHeader, suite SecuritySuite) ([]byte, error) {
	switch suite {
	case SecuritySuite0:
		return encryptGCM(key, plaintext, serverSystemTitle, header)
	case SecuritySuite1, SecuritySuite2:
		return encryptCBCandGMAC(key, plaintext, serverSystemTitle, header)
	case SecuritySuite3:
		return encryptKuznCmac(key, plaintext, serverSystemTitle, header)
	default:
		return nil, fmt.Errorf("unsupported security suite: %d", suite)
	}
}

// DecryptAndVerify decrypts and authenticates a ciphertext APDU.
func DecryptAndVerify(key, ciphertext, serverSystemTitle []byte, header *SecurityHeader, suite SecuritySuite, lastFrameCounter uint32) ([]byte, error) {
	switch suite {
	case SecuritySuite0:
		return decryptGCM(key, ciphertext, serverSystemTitle, header, lastFrameCounter)
	case SecuritySuite1, SecuritySuite2:
		return decryptCBCandGMAC(key, ciphertext, serverSystemTitle, header, lastFrameCounter)
	case SecuritySuite3:
		return decryptKuznCmac(key, ciphertext, serverSystemTitle, header, lastFrameCounter)
	default:
		return nil, fmt.Errorf("unsupported security suite: %d", suite)
	}
}

func encryptGCM(key, plaintext, serverSystemTitle []byte, header *SecurityHeader) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := makeGCMNonce(serverSystemTitle, header.FrameCounter)
	additionalData, err := header.Encode()
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, additionalData)

	return ciphertext, nil
}

func decryptGCM(key, ciphertext, serverSystemTitle []byte, header *SecurityHeader, lastFrameCounter uint32) ([]byte, error) {
	if header.FrameCounter <= lastFrameCounter {
		return nil, ErrReplayAttack
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := makeGCMNonce(serverSystemTitle, header.FrameCounter)
	additionalData, err := header.Encode()
	if err != nil {
		return nil, err
	}
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}

	return plaintext, nil
}

func encryptCBCandGMAC(key, plaintext, serverSystemTitle []byte, header *SecurityHeader) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := makeCBCIV(block, serverSystemTitle, header.FrameCounter)

	// Encrypt
	paddedPlaintext, err := pkcs7Pad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// Authenticate
	headerBytes, err := header.Encode()
	if err != nil {
		return nil, err
	}
	authenticatedData := make([]byte, len(headerBytes)+len(ciphertext))
	copy(authenticatedData, headerBytes)
	copy(authenticatedData[len(headerBytes):], ciphertext)

	nonce := makeGCMNonce(serverSystemTitle, header.FrameCounter)
	tag, err := gmac(key, nonce, authenticatedData)
	if err != nil {
		return nil, err
	}

	return append(ciphertext, tag...), nil
}

func decryptCBCandGMAC(key, ciphertext, serverSystemTitle []byte, header *SecurityHeader, lastFrameCounter uint32) ([]byte, error) {
	if header.FrameCounter <= lastFrameCounter {
		return nil, ErrReplayAttack
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := makeCBCIV(block, serverSystemTitle, header.FrameCounter)

	// Verify tag
	if len(ciphertext) < 12 {
		return nil, ErrAuthenticationFailed
	}
	tag := ciphertext[len(ciphertext)-12:]
	ciphertext = ciphertext[:len(ciphertext)-12]
	headerBytes, err := header.Encode()
	if err != nil {
		return nil, err
	}
	authenticatedData := make([]byte, len(headerBytes)+len(ciphertext))
	copy(authenticatedData, headerBytes)
	copy(authenticatedData[len(headerBytes):], ciphertext)

	nonce := makeGCMNonce(serverSystemTitle, header.FrameCounter)
	expectedTag, err := gmac(key, nonce, authenticatedData)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(tag, expectedTag) {
		return nil, ErrAuthenticationFailed
	}

	// Decrypt
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	unpaddedPlaintext, err := pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	return unpaddedPlaintext, nil
}

func makeGCMNonce(systemTitle []byte, frameCounter uint32) []byte {
	nonce := make([]byte, 12)
	copy(nonce, systemTitle)
	nonce[8] = byte(frameCounter >> 24)
	nonce[9] = byte(frameCounter >> 16)
	nonce[10] = byte(frameCounter >> 8)
	nonce[11] = byte(frameCounter)
	return nonce
}

func makeCBCIV(block cipher.Block, systemTitle []byte, frameCounter uint32) []byte {
	ivSource := make([]byte, 16)
	copy(ivSource, systemTitle)
	ivSource[8] = byte(frameCounter >> 24)
	ivSource[9] = byte(frameCounter >> 16)
	ivSource[10] = byte(frameCounter >> 8)
	ivSource[11] = byte(frameCounter)
	iv := make([]byte, aes.BlockSize)
	block.Encrypt(iv, ivSource)
	return iv
}

func gmac(key, nonce, authenticatedData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithTagSize(block, 12)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size: got %d, want %d", len(nonce), gcm.NonceSize())
	}

	tag := gcm.Seal(nil, nonce, nil, authenticatedData)
	return tag, nil
}

func pkcs7Pad(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, fmt.Errorf("invalid block size")
	}
	padding := blockSize - (len(data) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...), nil
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, fmt.Errorf("invalid block size")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("invalid data length")
	}
	padding := int(data[len(data)-1])
	if padding > blockSize || padding == 0 {
		return nil, ErrInvalidPadding
	}
	pad := data[len(data)-padding:]
	for i := 0; i < padding; i++ {
		if pad[i] != byte(padding) {
			return nil, ErrInvalidPadding
		}
	}
	return data[:len(data)-padding], nil
}
