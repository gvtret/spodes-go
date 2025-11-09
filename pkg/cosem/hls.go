package cosem

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// APDUType constants for secured APDUs.
const (
	APDU_GLO_GET_REQUEST  APDUType = 0xC8
	APDU_GLO_SET_REQUEST  APDUType = 0xC9
	APDU_GLO_ACTION_REQUEST APDUType = 0xCB
	APDU_GLO_GET_RESPONSE APDUType = 0xCC
	APDU_GLO_SET_RESPONSE APDUType = 0xCD
	APDU_GLO_ACTION_RESPONSE APDUType = 0xCF
)

// SecurityControl byte flags.
const (
	SecurityControlAuthenticationOnly SecurityControl = 0x10
	SecurityControlEncryptionOnly     SecurityControl = 0x20
	SecurityControlAuthenticatedAndEncrypted SecurityControl = 0x30
)

// SecurityControl represents the security control byte.
type SecurityControl byte

// SecurityHeader represents the security header of a secured APDU.
type SecurityHeader struct {
	SecurityControl    SecurityControl
	FrameCounter       uint32
	AuthenticatedData []byte // Additional authenticated data
}

// ErrReplayAttack is returned when a replay attack is detected.
var ErrReplayAttack = fmt.Errorf("replay attack detected")

// ErrAuthenticationFailed is returned when authentication fails.
var ErrAuthenticationFailed = fmt.Errorf("authentication failed")

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
func EncryptAndTag(key, plaintext, serverSystemTitle []byte, header *SecurityHeader) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// The IV is the system title (8 bytes) followed by the frame counter (4 bytes).
	iv := make([]byte, 12)
	copy(iv, serverSystemTitle)
	iv[8] = byte(header.FrameCounter >> 24)
	iv[9] = byte(header.FrameCounter >> 16)
	iv[10] = byte(header.FrameCounter >> 8)
	iv[11] = byte(header.FrameCounter)

	additionalData, err := header.Encode()
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, iv, plaintext, additionalData)

	return ciphertext, nil
}

// DecryptAndVerify decrypts and authenticates a ciphertext APDU.
func DecryptAndVerify(key, ciphertext, serverSystemTitle []byte, header *SecurityHeader, lastFrameCounter uint32) ([]byte, error) {
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

	// The IV is the system title (8 bytes) followed by the frame counter (4 bytes).
	iv := make([]byte, 12)
	copy(iv, serverSystemTitle)
	iv[8] = byte(header.FrameCounter >> 24)
	iv[9] = byte(header.FrameCounter >> 16)
	iv[10] = byte(header.FrameCounter >> 8)
	iv[11] = byte(header.FrameCounter)

	additionalData, err := header.Encode()
	if err != nil {
		return nil, err
	}
	plaintext, err := aesgcm.Open(nil, iv, ciphertext, additionalData)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}

	return plaintext, nil
}
