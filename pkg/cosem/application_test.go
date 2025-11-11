package cosem

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApplication_HandleGetRequest(t *testing.T) {
	obisAssociationLN, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obisAssociationLN)
	obisSecurity, _ := NewObisCodeFromString("0.0.43.0.0.255")
	clientSystemTitle := []byte("CLIENT")
	serverSystemTitle := []byte("SERVER")
	masterKey := []byte("master_key")
	guek := []byte("global_unicast_key")
	gak := []byte("global_auth_key")
	securitySetup, _ := NewSecuritySetup(*obisSecurity, clientSystemTitle, serverSystemTitle, masterKey, guek, gak)

	app := NewApplication(nil, associationLN, securitySetup)

	obis, _ := NewObisCodeFromString("1.0.0.3.0.255")
	dataObj, _ := NewData(*obis, uint32(12345))
	app.RegisterObject(dataObj)

	t.Run("Successful Get", func(t *testing.T) {
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
		encodedResp, err := app.HandleAPDU(encodedReq)
		if err != nil {
			t.Fatalf("HandleAPDU failed: %v", err)
		}
		resp := &GetResponse{}
		err = resp.Decode(encodedResp)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if resp.Result.IsDataAccessResult {
			t.Fatalf("Expected data, got DataAccessResult: %v", resp.Result.Value)
		}

		if val, ok := resp.Result.Value.(uint32); !ok || val != 12345 {
			t.Errorf("Value mismatch: got %v, want %v", resp.Result.Value, 12345)
		}
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
		encodedResp, err := app.HandleAPDU(encodedReq)
		if err != nil {
			t.Fatalf("HandleAPDU failed: %v", err)
		}
		resp := &GetResponse{}
		err = resp.Decode(encodedResp)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if !resp.Result.IsDataAccessResult {
			t.Fatal("Expected DataAccessResult, got data")
		}

		if resp.Result.Value.(DataAccessResultEnum) != OBJECT_UNDEFINED {
			t.Errorf("DataAccessResult mismatch: got %v, want %v", resp.Result.Value, OBJECT_UNDEFINED)
		}
	})
}

func TestApplication_HandleSetRequest(t *testing.T) {
	obisAssociationLN, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obisAssociationLN)
	obisSecurity, _ := NewObisCodeFromString("0.0.43.0.0.255")
	clientSystemTitle := []byte("CLIENT")
	serverSystemTitle := []byte("SERVER")
	masterKey := []byte("master_key")
	guek := []byte("global_unicast_key")
	gak := []byte("global_auth_key")
	securitySetup, _ := NewSecuritySetup(*obisSecurity, clientSystemTitle, serverSystemTitle, masterKey, guek, gak)
	app := NewApplication(nil, associationLN, securitySetup)

	obis, _ := NewObisCodeFromString("1.0.0.3.0.255")
	dataObj, _ := NewData(*obis, uint32(12345))
	app.RegisterObject(dataObj)

	t.Run("Successful Set", func(t *testing.T) {
		req := &SetRequest{
			Type:                SET_REQUEST_NORMAL,
			InvokeIDAndPriority: 0x81,
			AttributeDescriptor: CosemAttributeDescriptor{
				ClassID:     DataClassID,
				InstanceID:  *obis,
				AttributeID: 2,
			},
			Value: uint32(54321),
		}

		encodedReq, _ := req.Encode()
		encodedResp, err := app.HandleAPDU(encodedReq)
		if err != nil {
			t.Fatalf("HandleAPDU failed: %v", err)
		}
		resp := &SetResponse{}
		err = resp.Decode(encodedResp)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if resp.Result != SUCCESS {
			t.Fatalf("Expected SUCCESS, got %v", resp.Result)
		}

		val, _ := dataObj.GetAttribute(2)
		if val.(uint32) != 54321 {
			t.Errorf("Value not set correctly: got %v, want %v", val, 54321)
		}
	})

	t.Run("Object Not Found", func(t *testing.T) {
		unknownObis, _ := NewObisCodeFromString("0.0.0.0.0.0")
		req := &SetRequest{
			Type:                SET_REQUEST_NORMAL,
			InvokeIDAndPriority: 0x81,
			AttributeDescriptor: CosemAttributeDescriptor{
				ClassID:     DataClassID,
				InstanceID:  *unknownObis,
				AttributeID: 2,
			},
			Value: uint32(54321),
		}

		encodedReq, _ := req.Encode()
		encodedResp, err := app.HandleAPDU(encodedReq)
		if err != nil {
			t.Fatalf("HandleAPDU failed: %v", err)
		}
		resp := &SetResponse{}
		err = resp.Decode(encodedResp)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if resp.Result != OBJECT_UNDEFINED {
			t.Errorf("Expected OBJECT_UNDEFINED, got %v", resp.Result)
		}
	})
}

func TestApplication_HandleActionRequest(t *testing.T) {
	obisAssociationLN, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obisAssociationLN)
	obisSecurity, _ := NewObisCodeFromString("0.0.43.0.0.255")
	clientSystemTitle := []byte("CLIENT")
	serverSystemTitle := []byte("SERVER")
	masterKey := []byte("master_key")
	guek := []byte("global_unicast_key")
	gak := []byte("global_auth_key")
	securitySetup, _ := NewSecuritySetup(*obisSecurity, clientSystemTitle, serverSystemTitle, masterKey, guek, gak)
	app := NewApplication(nil, associationLN, securitySetup)

	obis, _ := NewObisCodeFromString("1.0.0.4.0.255")
	scalerUnit := ScalerUnit{Scaler: 0, Unit: UnitCount}
	registerObj, _ := NewRegister(*obis, int32(100), scalerUnit)
	app.RegisterObject(registerObj)

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
		encodedResp, err := app.HandleAPDU(encodedReq)
		if err != nil {
			t.Fatalf("HandleAPDU failed: %v", err)
		}
		resp := &ActionResponse{}
		err = resp.Decode(encodedResp)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if resp.Result.IsDataAccessResult {
			t.Fatalf("Expected data, got DataAccessResult: %v", resp.Result.Value)
		}

		val, _ := registerObj.GetAttribute(2)
		if val.(int32) != 0 {
			t.Errorf("Value not reset correctly: got %v, want %v", val, 0)
		}
	})

	t.Run("Object Not Found", func(t *testing.T) {
		unknownObis, _ := NewObisCodeFromString("0.0.0.0.0.0")
		req := &ActionRequest{
			Type:                ACTION_REQUEST_NORMAL,
			InvokeIDAndPriority: 0x81,
			MethodDescriptor: CosemMethodDescriptor{
				ClassID:    RegisterClassID,
				InstanceID: *unknownObis,
				MethodID:   1, // reset method
			},
			Parameters: []interface{}{},
		}

		encodedReq, _ := req.Encode()
		encodedResp, err := app.HandleAPDU(encodedReq)
		if err != nil {
			t.Fatalf("HandleAPDU failed: %v", err)
		}
		resp := &ActionResponse{}
		err = resp.Decode(encodedResp)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if !resp.Result.IsDataAccessResult {
			t.Fatal("Expected DataAccessResult, got data")
		}

		if resp.Result.Value.(DataAccessResultEnum) != OBJECT_UNDEFINED {
			t.Errorf("DataAccessResult mismatch: got %v, want %v", resp.Result.Value, OBJECT_UNDEFINED)
		}
	})
}

func TestApplication_SecurityPolicy(t *testing.T) {
	obisAssociationLN, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obisAssociationLN)
	obisSecurity, _ := NewObisCodeFromString("0.0.43.0.0.255")
	clientSystemTitle := []byte("CLIENT")
	serverSystemTitle := []byte("SERVER01")
	masterKey := []byte("master_key")
	guek := []byte("0123456789ABCDEF")
	gak := []byte("0123456789ABCDEF")
	securitySetup, _ := NewSecuritySetup(*obisSecurity, clientSystemTitle, serverSystemTitle, masterKey, guek, gak)
	app := NewApplication(nil, associationLN, securitySetup)
	app.securitySetup.SetAttribute(2, PolicyAuthenticatedRequest)

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

	_, err := app.HandleAPDU(encodedReq)
	assert.Error(t, err)

	header := &SecurityHeader{
		SecurityControl: SecurityControlEncryptionOnly, // Policy requires authentication
		FrameCounter:    1,
	}

	ciphertext, err := EncryptAndTag(guek, encodedReq, serverSystemTitle, header, SecuritySuite0)
	assert.NoError(t, err)
	encodedHeader, _ := header.Encode()
	securedReq := append([]byte{byte(APDU_GLO_GET_REQUEST)}, append(encodedHeader, ciphertext...)...)

	_, err = app.HandleAPDU(securedReq)
	assert.Error(t, err)
}
