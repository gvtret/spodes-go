package cosem

import (
	"crypto/cipher"
	"crypto/subtle"
	"fmt"
	"github.com/aead/cmac"
	"github.com/ddulesov/gogost/gost34112012256"
	"github.com/ddulesov/gogost/gost3412128"
)

const (
	kuznyechikBlockSize = 16
)

func DeriveKeys(masterKey, context []byte) ([]byte, []byte) {
	h := gost34112012256.New()
	h.Write(append(append([]byte("DLMS-KUZ-ENC"), masterKey...), context...))
	ke := h.Sum(nil)
	h.Reset()
	h.Write(append(append([]byte("DLMS-KUZ-AUTH"), masterKey...), context...))
	ka := h.Sum(nil)
	return ke, ka
}

func ctrEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key size for Kuznyechik")
	}
	if len(iv) != kuznyechikBlockSize {
		return nil, fmt.Errorf("invalid IV size for CTR mode")
	}

	block := gost3412128.NewCipher(key)

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)
	return ciphertext, nil
}

func encryptKuznCmac(key, plaintext, serverSystemTitle []byte, header *SecurityHeader) ([]byte, error) {
	additionalData, err := header.Encode()
	if err != nil {
		return nil, err
	}

	iv := make([]byte, kuznyechikBlockSize)
	copy(iv, serverSystemTitle)
	iv[8] = byte(header.FrameCounter >> 24)
	iv[9] = byte(header.FrameCounter >> 16)
	iv[10] = byte(header.FrameCounter >> 8)
	iv[11] = byte(header.FrameCounter)

	context := append(serverSystemTitle, byte(SecuritySuite3))
	ke, ka := DeriveKeys(key, context)

	ciphertext, err := ctrEncrypt(ke, iv, plaintext)
	if err != nil {
		return nil, err
	}

	block := gost3412128.NewCipher(ka)
	tag, err := cmac.Sum(append(additionalData, ciphertext...), block, 16)
	if err != nil {
		return nil, err
	}

	return append(ciphertext, tag...), nil
}

func decryptKuznCmac(key, ciphertext, serverSystemTitle []byte, header *SecurityHeader, lastFrameCounter uint32) ([]byte, error) {
	if header.FrameCounter <= lastFrameCounter {
		return nil, ErrReplayAttack
	}

	tag := ciphertext[len(ciphertext)-kuznyechikBlockSize:]
	ciphertext = ciphertext[:len(ciphertext)-kuznyechikBlockSize]

	additionalData, err := header.Encode()
	if err != nil {
		return nil, err
	}

	context := append(serverSystemTitle, byte(SecuritySuite3))
	ke, ka := DeriveKeys(key, context)

	block := gost3412128.NewCipher(ka)
	expectedTag, err := cmac.Sum(append(additionalData, ciphertext...), block, 16)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare(tag, expectedTag) != 1 {
		return nil, ErrAuthenticationFailed
	}

	iv := make([]byte, kuznyechikBlockSize)
	copy(iv, serverSystemTitle)
	iv[8] = byte(header.FrameCounter >> 24)
	iv[9] = byte(header.FrameCounter >> 16)
	iv[10] = byte(header.FrameCounter >> 8)
	iv[11] = byte(header.FrameCounter)
	plaintext, err := ctrEncrypt(ke, iv, ciphertext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
