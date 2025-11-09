package cosem

import (
	"testing"
)

func TestApplication_HandleGetRequest(t *testing.T) {
	app := NewApplication()

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

		resp := app.HandleGetRequest(req)

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

		resp := app.HandleGetRequest(req)

		if !resp.Result.IsDataAccessResult {
			t.Fatal("Expected DataAccessResult, got data")
		}

		if resp.Result.Value.(DataAccessResultEnum) != OBJECT_UNDEFINED {
			t.Errorf("DataAccessResult mismatch: got %v, want %v", resp.Result.Value, OBJECT_UNDEFINED)
		}
	})
}

func TestApplication_HandleSetRequest(t *testing.T) {
	app := NewApplication()

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

		resp := app.HandleSetRequest(req)

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

		resp := app.HandleSetRequest(req)

		if resp.Result != OBJECT_UNDEFINED {
			t.Errorf("Expected OBJECT_UNDEFINED, got %v", resp.Result)
		}
	})
}

func TestApplication_HandleActionRequest(t *testing.T) {
	app := NewApplication()

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

		resp := app.HandleActionRequest(req)

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

		resp := app.HandleActionRequest(req)

		if !resp.Result.IsDataAccessResult {
			t.Fatal("Expected DataAccessResult, got data")
		}

		if resp.Result.Value.(DataAccessResultEnum) != OBJECT_UNDEFINED {
			t.Errorf("DataAccessResult mismatch: got %v, want %v", resp.Result.Value, OBJECT_UNDEFINED)
		}
	})
}
