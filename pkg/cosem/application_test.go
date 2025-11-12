package cosem

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAddr string

func (a mockAddr) Network() string { return "mock" }
func (a mockAddr) String() string  { return string(a) }

func setupTestApp(t *testing.T) (*Application, *AssociationLN, net.Addr, *Data) {
	t.Helper()

	obisAssociationLN, err := NewObisCodeFromString("0.0.40.0.0.255")
	require.NoError(t, err)
	associationLN, err := NewAssociationLN(*obisAssociationLN)
	require.NoError(t, err)

	obisSecurity, err := NewObisCodeFromString("0.0.43.0.0.255")
	require.NoError(t, err)
	securitySetup, err := NewSecuritySetup(*obisSecurity, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	app := NewApplication(nil, securitySetup)

	clientAddr := mockAddr("client1")
	app.AddAssociation(clientAddr.String(), associationLN)

	obis, err := NewObisCodeFromString("1.0.0.3.0.255")
	require.NoError(t, err)
	dataObj, err := NewData(*obis, uint32(12345))
	require.NoError(t, err)
	app.RegisterObject(dataObj)

	err = app.PopulateObjectList(associationLN, []ObisCode{*obis})
	require.NoError(t, err)

	return app, associationLN, clientAddr, dataObj
}

func TestApplication_HandleGetRequest(t *testing.T) {
	app, _, clientAddr, dataObj := setupTestApp(t)

	t.Run("Successful Get", func(t *testing.T) {
		req := &GetRequest{
			Type:                GET_REQUEST_NORMAL,
			InvokeIDAndPriority: 0x81,
			AttributeDescriptor: CosemAttributeDescriptor{
				ClassID:     DataClassID,
				InstanceID:  dataObj.InstanceID,
				AttributeID: 2,
			},
		}

		encodedReq, _ := req.Encode()
		encodedResp, err := app.HandleAPDU(encodedReq, clientAddr)
		assert.NoError(t, err)

		resp := &GetResponse{}
		err = resp.Decode(encodedResp)
		assert.NoError(t, err)

		assert.False(t, resp.Result.IsDataAccessResult)
		val, ok := resp.Result.Value.(uint32)
		assert.True(t, ok)
		assert.Equal(t, uint32(12345), val)
	})

	t.Run("Object Not Found", func(t *testing.T) {
		unknownObis, _ := NewObisCodeFromString("0.0.0.0.0.0")
		req := &GetRequest{
			Type:                GET_REQUEST_NORMAL,
			InvokeIDAndPriority: 0x81,
			AttributeDescriptor: CosemAttributeDescriptor{
				ClassID:     DataClassID,
				InstanceID:  *unknownObis,
				AttributeID: 2,
			},
		}

		encodedReq, _ := req.Encode()
		encodedResp, err := app.HandleAPDU(encodedReq, clientAddr)
		assert.NoError(t, err)

		resp := &GetResponse{}
		err = resp.Decode(encodedResp)
		assert.NoError(t, err)

		assert.True(t, resp.Result.IsDataAccessResult)
		assert.Equal(t, READ_WRITE_DENIED, resp.Result.Value)
	})

	t.Run("Access Denied", func(t *testing.T) {
		// Create a new data object that is not in the association's object list
		otherObis, _ := NewObisCodeFromString("1.1.1.1.1.1")
		otherDataObj, _ := NewData(*otherObis, uint32(999))
		app.RegisterObject(otherDataObj)

		req := &GetRequest{
			Type:                GET_REQUEST_NORMAL,
			InvokeIDAndPriority: 0x81,
			AttributeDescriptor: CosemAttributeDescriptor{
				ClassID:     DataClassID,
				InstanceID:  *otherObis,
				AttributeID: 2,
			},
		}

		encodedReq, _ := req.Encode()
		encodedResp, err := app.HandleAPDU(encodedReq, clientAddr)
		assert.NoError(t, err)

		resp := &GetResponse{}
		err = resp.Decode(encodedResp)
		assert.NoError(t, err)

		assert.True(t, resp.Result.IsDataAccessResult)
		assert.Equal(t, READ_WRITE_DENIED, resp.Result.Value)
	})
}

func TestApplication_HandleSetRequest(t *testing.T) {
	app, _, clientAddr, dataObj := setupTestApp(t)

	t.Run("Successful Set", func(t *testing.T) {
		req := &SetRequest{
			Type:                SET_REQUEST_NORMAL,
			InvokeIDAndPriority: 0x81,
			AttributeDescriptor: CosemAttributeDescriptor{
				ClassID:     DataClassID,
				InstanceID:  dataObj.InstanceID,
				AttributeID: 2,
			},
			Value: uint32(54321),
		}

		encodedReq, _ := req.Encode()
		encodedResp, err := app.HandleAPDU(encodedReq, clientAddr)
		assert.NoError(t, err)

		resp := &SetResponse{}
		err = resp.Decode(encodedResp)
		assert.NoError(t, err)

		assert.Equal(t, SUCCESS, resp.Result)

		val, _ := dataObj.GetAttribute(2)
		assert.Equal(t, uint32(54321), val)
	})
}

func TestApplication_HandleActionRequest(t *testing.T) {
	app, assoc, clientAddr, _ := setupTestApp(t)

	obis, err := NewObisCodeFromString("1.0.0.4.0.255")
	require.NoError(t, err)
	scalerUnit := ScalerUnit{Scaler: 0, Unit: UnitCount}
	registerObj, err := NewRegister(*obis, int32(100), scalerUnit)
	require.NoError(t, err)
	app.RegisterObject(registerObj)
	err = app.PopulateObjectList(assoc, []ObisCode{*obis})
	require.NoError(t, err)

	t.Run("Successful Action", func(t *testing.T) {
		req := &ActionRequest{
			Type:                ACTION_REQUEST_NORMAL,
			InvokeIDAndPriority: 0x81,
			MethodDescriptor: CosemMethodDescriptor{
				ClassID:    RegisterClassID,
				InstanceID: *obis,
				MethodID:   1, // reset method
			},
			Parameters: []interface{}{},
		}

		encodedReq, _ := req.Encode()
		encodedResp, err := app.HandleAPDU(encodedReq, clientAddr)
		assert.NoError(t, err)

		resp := &ActionResponse{}
		err = resp.Decode(encodedResp)
		assert.NoError(t, err)

		assert.False(t, resp.Result.IsDataAccessResult)

		val, _ := registerObj.GetAttribute(2)
		assert.Equal(t, int32(0), val)
	})
}

func TestApplication_SecurityPolicy(t *testing.T) {
	app, _, clientAddr, dataObj := setupTestApp(t)
	err := app.securitySetup.SetAttribute(2, PolicyAuthenticatedRequest)
	require.NoError(t, err)

	req := &GetRequest{
		Type:                GET_REQUEST_NORMAL,
		InvokeIDAndPriority: 0x81,
		AttributeDescriptor: CosemAttributeDescriptor{
			ClassID:     DataClassID,
			InstanceID:  dataObj.InstanceID,
			AttributeID: 2,
		},
	}
	encodedReq, _ := req.Encode()

	_, err = app.HandleAPDU(encodedReq, clientAddr)
	assert.Error(t, err)

	header := &SecurityHeader{
		SecurityControl: SecurityControlEncryptionOnly, // Policy requires authentication
		FrameCounter:    1,
	}

	// Mocking encryption/decryption is complex, so we're just testing the policy check here.
	// We expect an error because the security control level is not sufficient.
	// This test doesn't actually try to decrypt the request.
	guek := []byte("0123456789ABCDEF")
	serverSystemTitle := []byte("SERVER01")
	ciphertext, err := EncryptAndTag(guek, encodedReq, serverSystemTitle, header, SecuritySuite0)
	assert.NoError(t, err)
	encodedHeader, _ := header.Encode()
	securedReq := append([]byte{byte(APDU_GLO_GET_REQUEST)}, append(encodedHeader, ciphertext...)...)

	_, err = app.HandleAPDU(securedReq, clientAddr)
	assert.Error(t, err)
}
