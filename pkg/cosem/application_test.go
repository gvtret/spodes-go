package cosem

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAddr string

func (a mockAddr) Network() string { return "mock" }
func (a mockAddr) String() string  { return string(a) }

func setupTestApp() (*Application, *AssociationLN, net.Addr, *Data) {
	obisAssociationLN, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obisAssociationLN)

	obisSecurity, _ := NewObisCodeFromString("0.0.43.0.0.255")
	securitySetup, _ := NewSecuritySetup(*obisSecurity, nil, nil, nil, nil, nil)

	app := NewApplication(nil, securitySetup)

	clientAddr := mockAddr("client1")
	app.AddAssociation(clientAddr.String(), associationLN)

	obis, _ := NewObisCodeFromString("1.0.0.3.0.255")
	dataObj, _ := NewData(*obis, uint32(12345))
	app.RegisterObject(dataObj)

	app.PopulateObjectList(associationLN, []ObisCode{*obis})

	return app, associationLN, clientAddr, dataObj
}

func TestApplication_HandleGetRequest(t *testing.T) {
	app, _, clientAddr, dataObj := setupTestApp()

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
	app, _, clientAddr, dataObj := setupTestApp()

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
	app, assoc, clientAddr, _ := setupTestApp()

	obis, _ := NewObisCodeFromString("1.0.0.4.0.255")
	scalerUnit := ScalerUnit{Scaler: 0, Unit: UnitCount}
	registerObj, _ := NewRegister(*obis, int32(100), scalerUnit)
	app.RegisterObject(registerObj)
	app.PopulateObjectList(assoc, []ObisCode{*obis})

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
	app, _, clientAddr, dataObj := setupTestApp()
	app.securitySetup.SetAttribute(2, PolicyAuthenticatedRequest)

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

	_, err := app.HandleAPDU(encodedReq, clientAddr)
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
