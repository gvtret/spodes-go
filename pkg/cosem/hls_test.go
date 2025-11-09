package cosem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptAndTag_DecryptAndVerify(t *testing.T) {
	key := []byte("0123456789ABCDEF")
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")
	header := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext, err := EncryptAndTag(key, plaintext, serverSystemTitle, header)
	assert.NoError(t, err)

	decrypted, err := DecryptAndVerify(key, ciphertext, serverSystemTitle, header, 0)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptAndVerify_ReplayAttack(t *testing.T) {
	key := []byte("0123456789ABCDEF")
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")
	header := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext, err := EncryptAndTag(key, plaintext, serverSystemTitle, header)
	assert.NoError(t, err)

	_, err = DecryptAndVerify(key, ciphertext, serverSystemTitle, header, 1)
	assert.Equal(t, ErrReplayAttack, err)
}

func TestApplication_HandleAPDU_Secured(t *testing.T) {
	obisAssociationLN, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obisAssociationLN)
	obisSecurity, _ := NewObisCodeFromString("0.0.43.0.0.255")
	clientSystemTitle := []byte("CLIENT")
	serverSystemTitle := []byte("SERVER01")
	masterKey := []byte("master_key")
	guek := []byte("0123456789ABCDEF")
	gak := []byte("0123456789ABCDEF")
	securitySetup, _ := NewSecuritySetup(*obisSecurity, clientSystemTitle, serverSystemTitle, masterKey, guek, gak)
	app := NewApplication(associationLN, securitySetup)

	obis, _ := NewObisCodeFromString("1.0.0.3.0.255")
	dataObj, _ := NewData(*obis, uint32(12345))
	app.RegisterObject(dataObj)

	req := &GetRequest{
		Type:                GET_REQUEST_NORMAL,
		InvokeIDAndPriority: 0x81,
		AttributeDescriptor: CosemAttributeDescriptor{
			ClassID:     DataClassID,
			InstanceID:  *obis,
			AttributeID: 2,
		},
	}
	encodedReq, _ := req.Encode()

	header := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext, err := EncryptAndTag(guek, encodedReq, serverSystemTitle, header)
	assert.NoError(t, err)
	encodedHeader, _ := header.Encode()
	securedReq := append([]byte{byte(APDU_GLO_GET_REQUEST)}, append(encodedHeader, ciphertext...)...)

	encodedResp, err := app.HandleAPDU(securedReq)
	assert.NoError(t, err)

	respHeader := &SecurityHeader{}
	err = respHeader.Decode(encodedResp[1:])
	assert.NoError(t, err)

	plaintext, err := DecryptAndVerify(guek, encodedResp[6:], serverSystemTitle, respHeader, 0)
	assert.NoError(t, err)

	resp := &GetResponse{}
	err = resp.Decode(plaintext)
	assert.NoError(t, err)

	assert.False(t, resp.Result.IsDataAccessResult)
	assert.Equal(t, uint32(12345), resp.Result.Value.(uint32))
}
