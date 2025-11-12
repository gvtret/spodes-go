package cosem

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptAndTag_DecryptAndVerify_Suite0(t *testing.T) {
	key := []byte("0123456789ABCDEF")
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")
	header := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext, err := EncryptAndTag(key, plaintext, serverSystemTitle, header, SecuritySuite0)
	assert.NoError(t, err)

	decrypted, err := DecryptAndVerify(key, ciphertext, serverSystemTitle, header, SecuritySuite0, 0)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptAndTag_DecryptAndVerify_Suite2(t *testing.T) {
	key := []byte("0123456789ABCDEF0123456789ABCDEF")
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")
	header := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext, err := EncryptAndTag(key, plaintext, serverSystemTitle, header, SecuritySuite2)
	assert.NoError(t, err)

	decrypted, err := DecryptAndVerify(key, ciphertext, serverSystemTitle, header, SecuritySuite2, 0)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptAndTag_DecryptAndVerify_Suite1(t *testing.T) {
	key := []byte("0123456789ABCDEF")
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")
	header := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext, err := EncryptAndTag(key, plaintext, serverSystemTitle, header, SecuritySuite1)
	assert.NoError(t, err)

	decrypted, err := DecryptAndVerify(key, ciphertext, serverSystemTitle, header, SecuritySuite1, 0)
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

	ciphertext, err := EncryptAndTag(key, plaintext, serverSystemTitle, header, SecuritySuite0)
	assert.NoError(t, err)

	_, err = DecryptAndVerify(key, ciphertext, serverSystemTitle, header, SecuritySuite0, 1)
	assert.Equal(t, ErrReplayAttack, err)
}

func TestEncryptCBCandGMAC_TagVariesWithFrameCounter(t *testing.T) {
	key := []byte("0123456789ABCDEF")
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")

	header1 := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext1, err := encryptCBCandGMAC(key, plaintext, serverSystemTitle, header1)
	require.NoError(t, err)

	header2 := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    2,
	}

	ciphertext2, err := encryptCBCandGMAC(key, plaintext, serverSystemTitle, header2)
	require.NoError(t, err)

	tag1 := ciphertext1[len(ciphertext1)-12:]
	tag2 := ciphertext2[len(ciphertext2)-12:]

	assert.NotEqual(t, tag1, tag2, "tags should differ when frame counter changes")
}

func TestEncryptCBCandGMAC_KnownGoodTag(t *testing.T) {
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")

	testCases := []struct {
		name        string
		key         []byte
		suite       SecuritySuite
		expectedTag string
	}{
		{
			name:        "suite1",
			key:         []byte("0123456789ABCDEF"),
			suite:       SecuritySuite1,
			expectedTag: "041580fc210fca507be5471a",
		},
		{
			name:        "suite2",
			key:         []byte("0123456789ABCDEF0123456789ABCDEF"),
			suite:       SecuritySuite2,
			expectedTag: "19cfec5e5a42efdb631d9c01",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			header := &SecurityHeader{
				SecurityControl: SecurityControlAuthenticatedAndEncrypted,
				FrameCounter:    1,
			}

			ciphertext, err := EncryptAndTag(tc.key, plaintext, serverSystemTitle, header, tc.suite)
			require.NoError(t, err)

			require.GreaterOrEqual(t, len(ciphertext), 12)
			expectedTag, err := hex.DecodeString(tc.expectedTag)
			require.NoError(t, err)

			assert.Equal(t, expectedTag, ciphertext[len(ciphertext)-12:])

			decrypted, err := DecryptAndVerify(tc.key, ciphertext, serverSystemTitle, header, tc.suite, 0)
			require.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		})
	}
}

func TestDecryptCBCandGMAC_DetectsTampering(t *testing.T) {
	key := []byte("0123456789ABCDEF")
	plaintext := []byte("Hello, COSEM!")
	serverSystemTitle := []byte("SERVER01")
	header := &SecurityHeader{
		SecurityControl: SecurityControlAuthenticatedAndEncrypted,
		FrameCounter:    1,
	}

	ciphertext, err := encryptCBCandGMAC(key, plaintext, serverSystemTitle, header)
	require.NoError(t, err)

	t.Run("frame counter tampering", func(t *testing.T) {
		tamperedHeader := &SecurityHeader{
			SecurityControl: SecurityControlAuthenticatedAndEncrypted,
			FrameCounter:    2,
		}

		_, err := decryptCBCandGMAC(key, ciphertext, serverSystemTitle, tamperedHeader, 0)
		assert.ErrorIs(t, err, ErrAuthenticationFailed)
	})

	t.Run("ciphertext tampering", func(t *testing.T) {
		tamperedCiphertext := append([]byte(nil), ciphertext...)
		tamperedCiphertext[0] ^= 0xFF

		_, err := decryptCBCandGMAC(key, tamperedCiphertext, serverSystemTitle, header, 0)
		assert.ErrorIs(t, err, ErrAuthenticationFailed)
	})
}

func TestApplication_HandleAPDU_Secured(t *testing.T) {
	obisAssociationLN, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obisAssociationLN)
	obisSecurity, err := NewObisCodeFromString("0.0.43.0.0.255")
	require.NoError(t, err)
	clientSystemTitle := []byte("CLIENT")
	serverSystemTitle := []byte("SERVER01")
	masterKey := []byte("master_key")
	guek := []byte("0123456789ABCDEF")
	gak := []byte("0123456789ABCDEF")
	securitySetup, err := NewSecuritySetup(*obisSecurity, clientSystemTitle, serverSystemTitle, masterKey, guek, gak)
	require.NoError(t, err)
	app := NewApplication(nil, securitySetup)
	err = app.securitySetup.SetAttribute(3, SecuritySuite1)
	require.NoError(t, err)

	clientAddr := mockAddr("client1")
	app.AddAssociation(clientAddr.String(), associationLN)

	obis, err := NewObisCodeFromString("1.0.0.3.0.255")
	require.NoError(t, err)
	dataObj, err := NewData(*obis, uint32(12345))
	require.NoError(t, err)
	app.RegisterObject(dataObj)
	err = app.PopulateObjectList(associationLN, []ObisCode{*obis})
	require.NoError(t, err)

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

	buildRequest := func(counter uint32) []byte {
		header := &SecurityHeader{
			SecurityControl: SecurityControlAuthenticatedAndEncrypted,
			FrameCounter:    counter,
		}

		ciphertext, err := EncryptAndTag(guek, encodedReq, serverSystemTitle, header, SecuritySuite1)
		assert.NoError(t, err)
		encodedHeader, _ := header.Encode()
		return append([]byte{byte(APDU_GLO_GET_REQUEST)}, append(encodedHeader, ciphertext...)...)
	}

	lastServerCounter := associationLN.ServerInvocationCounter()

	for i := uint32(1); i <= 2; i++ {
		securedReq := buildRequest(i)

		encodedResp, err := app.HandleAPDU(securedReq, clientAddr)
		assert.NoError(t, err)

		respHeader := &SecurityHeader{}
		err = respHeader.Decode(encodedResp[1:])
		assert.NoError(t, err)
		assert.Equal(t, lastServerCounter+1, respHeader.FrameCounter)

		plaintext, err := DecryptAndVerify(guek, encodedResp[6:], serverSystemTitle, respHeader, SecuritySuite1, lastServerCounter)
		assert.NoError(t, err)

		lastServerCounter = respHeader.FrameCounter

		resp := &GetResponse{}
		err = resp.Decode(plaintext)
		assert.NoError(t, err)

		assert.False(t, resp.Result.IsDataAccessResult)
		assert.Equal(t, uint32(12345), resp.Result.Value.(uint32))
	}
}
