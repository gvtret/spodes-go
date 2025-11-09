package cosem

import (
	"testing"
)

func TestGetRequest(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.0.3.0.255")
	req := &GetRequest{
		Type:                GET_REQUEST_NORMAL,
		InvokeIDAndPriority: 0x81,
		AttributeDescriptor: CosemAttributeDescriptor{
			ClassID:     1,
			InstanceID:  *obis,
			AttributeID: 2,
		},
	}

	encoded, err := req.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded := &GetRequest{}
	err = decoded.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if req.Type != decoded.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, req.Type)
	}

	if req.InvokeIDAndPriority != decoded.InvokeIDAndPriority {
		t.Errorf("InvokeIDAndPriority mismatch: got %v, want %v", decoded.InvokeIDAndPriority, req.InvokeIDAndPriority)
	}

	if req.AttributeDescriptor.ClassID != decoded.AttributeDescriptor.ClassID {
		t.Errorf("ClassID mismatch: got %v, want %v", decoded.AttributeDescriptor.ClassID, req.AttributeDescriptor.ClassID)
	}

	if req.AttributeDescriptor.InstanceID.String() != decoded.AttributeDescriptor.InstanceID.String() {
		t.Errorf("InstanceID mismatch: got %v, want %v", decoded.AttributeDescriptor.InstanceID.String(), req.AttributeDescriptor.InstanceID.String())
	}

	if req.AttributeDescriptor.AttributeID != decoded.AttributeDescriptor.AttributeID {
		t.Errorf("AttributeID mismatch: got %v, want %v", decoded.AttributeDescriptor.AttributeID, req.AttributeDescriptor.AttributeID)
	}
}

func TestSetRequest(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.0.3.0.255")
	req := &SetRequest{
		Type:                SET_REQUEST_NORMAL,
		InvokeIDAndPriority: 0x81,
		AttributeDescriptor: CosemAttributeDescriptor{
			ClassID:     1,
			InstanceID:  *obis,
			AttributeID: 2,
		},
		Value: uint32(12345),
	}

	encoded, err := req.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded := &SetRequest{}
	err = decoded.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if req.Type != decoded.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, req.Type)
	}

	if req.InvokeIDAndPriority != decoded.InvokeIDAndPriority {
		t.Errorf("InvokeIDAndPriority mismatch: got %v, want %v", decoded.InvokeIDAndPriority, req.InvokeIDAndPriority)
	}

	if req.AttributeDescriptor.ClassID != decoded.AttributeDescriptor.ClassID {
		t.Errorf("ClassID mismatch: got %v, want %v", decoded.AttributeDescriptor.ClassID, req.AttributeDescriptor.ClassID)
	}

	if req.AttributeDescriptor.InstanceID.String() != decoded.AttributeDescriptor.InstanceID.String() {
		t.Errorf("InstanceID mismatch: got %v, want %v", decoded.AttributeDescriptor.InstanceID.String(), req.AttributeDescriptor.InstanceID.String())
	}

	if req.AttributeDescriptor.AttributeID != decoded.AttributeDescriptor.AttributeID {
		t.Errorf("AttributeID mismatch: got %v, want %v", decoded.AttributeDescriptor.AttributeID, req.AttributeDescriptor.AttributeID)
	}

	if req.Value.(uint32) != decoded.Value.(uint32) {
		t.Errorf("Value mismatch: got %v, want %v", decoded.Value, req.Value)
	}
}

func TestActionRequest(t *testing.T) {
	obis, _ := NewObisCodeFromString("1.0.0.4.0.255")
	req := &ActionRequest{
		Type:                ACTION_REQUEST_NORMAL,
		InvokeIDAndPriority: 0x81,
		MethodDescriptor: CosemMethodDescriptor{
			ClassID:    3,
			InstanceID: *obis,
			MethodID:   1,
		},
		Parameters: []interface{}{},
	}

	encoded, err := req.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded := &ActionRequest{}
	err = decoded.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if req.Type != decoded.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, req.Type)
	}

	if req.InvokeIDAndPriority != decoded.InvokeIDAndPriority {
		t.Errorf("InvokeIDAndPriority mismatch: got %v, want %v", decoded.InvokeIDAndPriority, req.InvokeIDAndPriority)
	}

	if req.MethodDescriptor.ClassID != decoded.MethodDescriptor.ClassID {
		t.Errorf("ClassID mismatch: got %v, want %v", decoded.MethodDescriptor.ClassID, req.MethodDescriptor.ClassID)
	}

	if req.MethodDescriptor.InstanceID.String() != decoded.MethodDescriptor.InstanceID.String() {
		t.Errorf("InstanceID mismatch: got %v, want %v", decoded.MethodDescriptor.InstanceID.String(), req.MethodDescriptor.InstanceID.String())
	}

	if req.MethodDescriptor.MethodID != decoded.MethodDescriptor.MethodID {
		t.Errorf("MethodID mismatch: got %v, want %v", decoded.MethodDescriptor.MethodID, req.MethodDescriptor.MethodID)
	}
}

func TestGetResponse_Data(t *testing.T) {
	resp := &GetResponse{
		Type:                GET_RESPONSE_NORMAL,
		InvokeIDAndPriority: 0x81,
		Result: GetDataResult{
			IsDataAccessResult: false,
			Value:              uint32(12345),
		},
	}

	encoded, err := resp.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded := &GetResponse{}
	err = decoded.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if resp.Type != decoded.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, resp.Type)
	}

	if resp.InvokeIDAndPriority != decoded.InvokeIDAndPriority {
		t.Errorf("InvokeIDAndPriority mismatch: got %v, want %v", decoded.InvokeIDAndPriority, resp.InvokeIDAndPriority)
	}

	if resp.Result.IsDataAccessResult != decoded.Result.IsDataAccessResult {
		t.Errorf("IsDataAccessResult mismatch: got %v, want %v", decoded.Result.IsDataAccessResult, resp.Result.IsDataAccessResult)
	}

	if resp.Result.Value.(uint32) != decoded.Result.Value.(uint32) {
		t.Errorf("Value mismatch: got %v, want %v", decoded.Result.Value, resp.Result.Value)
	}
}

func TestGetResponse_DataAccessResult(t *testing.T) {
	resp := &GetResponse{
		Type:                GET_RESPONSE_NORMAL,
		InvokeIDAndPriority: 0x81,
		Result: GetDataResult{
			IsDataAccessResult: true,
			Value:              OBJECT_UNDEFINED,
		},
	}

	encoded, err := resp.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded := &GetResponse{}
	err = decoded.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if resp.Type != decoded.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, resp.Type)
	}

	if resp.InvokeIDAndPriority != decoded.InvokeIDAndPriority {
		t.Errorf("InvokeIDAndPriority mismatch: got %v, want %v", decoded.InvokeIDAndPriority, resp.InvokeIDAndPriority)
	}

	if resp.Result.IsDataAccessResult != decoded.Result.IsDataAccessResult {
		t.Errorf("IsDataAccessResult mismatch: got %v, want %v", decoded.Result.IsDataAccessResult, resp.Result.IsDataAccessResult)
	}

	if resp.Result.Value.(DataAccessResultEnum) != decoded.Result.Value.(DataAccessResultEnum) {
		t.Errorf("Value mismatch: got %v, want %v", decoded.Result.Value, resp.Result.Value)
	}
}

func TestSetResponse(t *testing.T) {
	resp := &SetResponse{
		Type:                SET_RESPONSE_NORMAL,
		InvokeIDAndPriority: 0x81,
		Result:              SUCCESS,
	}

	encoded, err := resp.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded := &SetResponse{}
	err = decoded.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if resp.Type != decoded.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, resp.Type)
	}

	if resp.InvokeIDAndPriority != decoded.InvokeIDAndPriority {
		t.Errorf("InvokeIDAndPriority mismatch: got %v, want %v", decoded.InvokeIDAndPriority, resp.InvokeIDAndPriority)
	}

	if resp.Result != decoded.Result {
		t.Errorf("Result mismatch: got %v, want %v", decoded.Result, resp.Result)
	}
}

func TestActionResponse(t *testing.T) {
	resp := &ActionResponse{
		Type:                ACTION_RESPONSE_NORMAL,
		InvokeIDAndPriority: 0x81,
		Result: ActionResult{
			IsDataAccessResult: false,
			Value:              nil,
		},
	}

	encoded, err := resp.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded := &ActionResponse{}
	err = decoded.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if resp.Type != decoded.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, resp.Type)
	}

	if resp.InvokeIDAndPriority != decoded.InvokeIDAndPriority {
		t.Errorf("InvokeIDAndPriority mismatch: got %v, want %v", decoded.InvokeIDAndPriority, resp.InvokeIDAndPriority)
	}

	if resp.Result.IsDataAccessResult != decoded.Result.IsDataAccessResult {
		t.Errorf("IsDataAccessResult mismatch: got %v, want %v", decoded.Result.IsDataAccessResult, resp.Result.IsDataAccessResult)
	}

	if resp.Result.Value != decoded.Result.Value {
		t.Errorf("Value mismatch: got %v, want %v", decoded.Result.Value, resp.Result.Value)
	}
}
