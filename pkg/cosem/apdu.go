package cosem

import (
	"fmt"
	"github.com/gvtret/spodes-go/pkg/axdr"
)

// APDUType represents the type of an APDU.
type APDUType byte

const (
	APDU_GET_REQUEST     APDUType = 0xC0
	APDU_SET_REQUEST     APDUType = 0xC1
	APDU_ACTION_REQUEST  APDUType = 0xC3
	APDU_GET_RESPONSE    APDUType = 0xC4
	APDU_SET_RESPONSE    APDUType = 0xC5
	APDU_ACTION_RESPONSE APDUType = 0xC7
)

// GetRequestType represents the type of a Get-Request APDU.
type GetRequestType byte

const (
	GET_REQUEST_NORMAL GetRequestType = 0x01
)

// SetRequestType represents the type of a Set-Request APDU.
type SetRequestType byte

const (
	SET_REQUEST_NORMAL SetRequestType = 0x01
)

// ActionRequestType represents the type of an Action-Request APDU.
type ActionRequestType byte

const (
	ACTION_REQUEST_NORMAL ActionRequestType = 0x01
)

// GetResponseType represents the type of a Get-Response APDU.
type GetResponseType byte

const (
	GET_RESPONSE_NORMAL GetResponseType = 0x01
)

// SetResponseType represents the type of a Set-Response APDU.
type SetResponseType byte

const (
	SET_RESPONSE_NORMAL SetResponseType = 0x01
)

// ActionResponseType represents the type of an Action-Response APDU.
type ActionResponseType byte

const (
	ACTION_RESPONSE_NORMAL ActionResponseType = 0x01
)

// CosemAttributeDescriptor is the structure for a COSEM attribute descriptor.
type CosemAttributeDescriptor struct {
	ClassID     uint16
	InstanceID  ObisCode
	AttributeID int8
}

// CosemMethodDescriptor is the structure for a COSEM method descriptor.
type CosemMethodDescriptor struct {
	ClassID    uint16
	InstanceID ObisCode
	MethodID   int8
}

// GetRequest is the structure for a Get-Request APDU.
type GetRequest struct {
	Type                GetRequestType
	InvokeIDAndPriority uint8
	AttributeDescriptor CosemAttributeDescriptor
}

// SetRequest is the structure for a Set-Request APDU.
type SetRequest struct {
	Type                SetRequestType
	InvokeIDAndPriority uint8
	AttributeDescriptor CosemAttributeDescriptor
	Value               interface{}
}

// ActionRequest is the structure for an Action-Request APDU.
type ActionRequest struct {
	Type                ActionRequestType
	InvokeIDAndPriority uint8
	MethodDescriptor    CosemMethodDescriptor
	Parameters          interface{}
}

// GetResponse is the structure for a Get-Response APDU.
type GetResponse struct {
	Type                GetResponseType
	InvokeIDAndPriority uint8
	Result              GetDataResult
}

// SetResponse is the structure for a Set-Response APDU.
type SetResponse struct {
	Type                SetResponseType
	InvokeIDAndPriority uint8
	Result              DataAccessResultEnum
}

// ActionResponse is the structure for an Action-Response APDU.
type ActionResponse struct {
	Type                ActionResponseType
	InvokeIDAndPriority uint8
	Result              ActionResult
}

// GetDataResult represents the Get-Data-Result CHOICE.
// If IsDataAccessResult is true, Value holds a DataAccessResultEnum.
// If IsDataAccessResult is false, Value holds the requested data.
type GetDataResult struct {
	IsDataAccessResult bool
	Value              interface{}
}

// ActionResult represents the Action-Result CHOICE.
type ActionResult struct {
	IsDataAccessResult bool
	Value              interface{}
}

// Choice represents a CHOICE in A-XDR.
type Choice struct {
	Tag   uint8
	Value interface{}
}

// DataAccessResultEnum represents the result of a data access operation when it's not successful.
type DataAccessResultEnum byte

const (
	SUCCESS                   DataAccessResultEnum = 0
	HARDWARE_FAULT            DataAccessResultEnum = 1
	TEMPORARY_FAILURE         DataAccessResultEnum = 2
	READ_WRITE_DENIED         DataAccessResultEnum = 3
	OBJECT_UNDEFINED          DataAccessResultEnum = 4
	OBJECT_CLASS_INCONSISTENT DataAccessResultEnum = 9
	OBJECT_UNAVAILABLE        DataAccessResultEnum = 11
	TYPE_UNMATCHED            DataAccessResultEnum = 12
	SCOPE_OF_ACCESS_VIOLATED  DataAccessResultEnum = 13
	DATA_BLOCK_UNAVAILABLE    DataAccessResultEnum = 14
	LONG_GET_ABORTED          DataAccessResultEnum = 15
	NO_LONG_GET_IN_PROGRESS   DataAccessResultEnum = 16
	LONG_SET_ABORTED          DataAccessResultEnum = 17
	NO_LONG_SET_IN_PROGRESS   DataAccessResultEnum = 18
	DATA_BLOCK_NUMBER_INVALID DataAccessResultEnum = 19
	OTHER_REASON              DataAccessResultEnum = 250
)

// Encode encodes the GetRequest APDU into a byte slice.
func (gr *GetRequest) Encode() ([]byte, error) {
	obisBytes := gr.AttributeDescriptor.InstanceID.Bytes()
	attrDescStruct := axdr.Structure{
		gr.AttributeDescriptor.ClassID,
		obisBytes[:],
		gr.AttributeDescriptor.AttributeID,
	}

	reqStruct := axdr.Structure{
		byte(gr.Type),
		gr.InvokeIDAndPriority,
		attrDescStruct,
	}

	reqBody, err := axdr.Encode(reqStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to encode GetRequest body: %w", err)
	}

	res := append([]byte{byte(APDU_GET_REQUEST)}, reqBody...)
	return res, nil
}

// Decode decodes a byte slice into a GetRequest APDU.
func (gr *GetRequest) Decode(src []byte) error {
	if len(src) == 0 {
		return fmt.Errorf("empty source byte slice")
	}

	if APDUType(src[0]) != APDU_GET_REQUEST {
		return fmt.Errorf("invalid APDU tag for GetRequest: got %X, expected %X", src[0], APDU_GET_REQUEST)
	}

	decoded, err := axdr.Decode(src[1:])
	if err != nil {
		return err
	}
	decodedStruct, ok := decoded.(axdr.Structure)
	if !ok {
		return fmt.Errorf("expected a structure")
	}

	if len(decodedStruct) != 3 {
		return fmt.Errorf("expected 3 elements in GetRequest, got %d", len(decodedStruct))
	}

	reqType, ok := decodedStruct[0].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for GetRequestType")
	}
	gr.Type = GetRequestType(reqType)

	gr.InvokeIDAndPriority, ok = decodedStruct[1].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for InvokeIDAndPriority")
	}

	attrDescStruct, ok := decodedStruct[2].(axdr.Structure)
	if !ok {
		return fmt.Errorf("invalid type for AttributeDescriptor")
	}

	if len(attrDescStruct) != 3 {
		return fmt.Errorf("expected 3 elements in AttributeDescriptor, got %d", len(attrDescStruct))
	}

	classID, ok := attrDescStruct[0].(uint16)
	if !ok {
		return fmt.Errorf("invalid type for ClassID")
	}

	instanceID, ok := attrDescStruct[1].([]byte)
	if !ok {
		return fmt.Errorf("invalid type for InstanceID")
	}
	if len(instanceID) != 6 {
		return fmt.Errorf("invalid length for InstanceID: expected 6, got %d", len(instanceID))
	}

	attributeID, ok := attrDescStruct[2].(int8)
	if !ok {
		return fmt.Errorf("invalid type for AttributeID")
	}

	gr.AttributeDescriptor.ClassID = classID
	var instanceIDBytes [6]byte
	copy(instanceIDBytes[:], instanceID)
	gr.AttributeDescriptor.InstanceID.SetFromBytes(instanceIDBytes)
	gr.AttributeDescriptor.AttributeID = attributeID

	return nil
}

// Encode encodes the SetRequest APDU into a byte slice.
func (sr *SetRequest) Encode() ([]byte, error) {
	obisBytes := sr.AttributeDescriptor.InstanceID.Bytes()
	attrDescStruct := axdr.Structure{
		sr.AttributeDescriptor.ClassID,
		obisBytes[:],
		sr.AttributeDescriptor.AttributeID,
	}

	// The value is encoded with a data tag
	encodedValue, err := axdr.Encode(sr.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to encode value: %w", err)
	}

	reqStruct := axdr.Structure{
		byte(sr.Type),
		sr.InvokeIDAndPriority,
		attrDescStruct,
		encodedValue,
	}

	reqBody, err := axdr.Encode(reqStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to encode SetRequest body: %w", err)
	}

	res := append([]byte{byte(APDU_SET_REQUEST)}, reqBody...)
	return res, nil
}

// Decode decodes a byte slice into a SetRequest APDU.
func (sr *SetRequest) Decode(src []byte) error {
	if len(src) == 0 {
		return fmt.Errorf("empty source byte slice")
	}

	if APDUType(src[0]) != APDU_SET_REQUEST {
		return fmt.Errorf("invalid APDU tag for SetRequest: got %X, expected %X", src[0], APDU_SET_REQUEST)
	}

	decoded, err := axdr.Decode(src[1:])
	if err != nil {
		return err
	}
	decodedStruct, ok := decoded.(axdr.Structure)
	if !ok {
		return fmt.Errorf("expected a structure")
	}

	if len(decodedStruct) != 4 {
		return fmt.Errorf("expected 4 elements in SetRequest, got %d", len(decodedStruct))
	}

	reqType, ok := decodedStruct[0].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for SetRequestType")
	}
	sr.Type = SetRequestType(reqType)

	sr.InvokeIDAndPriority, ok = decodedStruct[1].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for InvokeIDAndPriority")
	}

	attrDescStruct, ok := decodedStruct[2].(axdr.Structure)
	if !ok {
		return fmt.Errorf("invalid type for AttributeDescriptor")
	}

	if len(attrDescStruct) != 3 {
		return fmt.Errorf("expected 3 elements in AttributeDescriptor, got %d", len(attrDescStruct))
	}

	classID, ok := attrDescStruct[0].(uint16)
	if !ok {
		return fmt.Errorf("invalid type for ClassID")
	}

	instanceID, ok := attrDescStruct[1].([]byte)
	if !ok {
		return fmt.Errorf("invalid type for InstanceID")
	}
	if len(instanceID) != 6 {
		return fmt.Errorf("invalid length for InstanceID: expected 6, got %d", len(instanceID))
	}

	attributeID, ok := attrDescStruct[2].(int8)
	if !ok {
		return fmt.Errorf("invalid type for AttributeID")
	}

	sr.AttributeDescriptor.ClassID = classID
	var instanceIDBytes [6]byte
	copy(instanceIDBytes[:], instanceID)
	sr.AttributeDescriptor.InstanceID.SetFromBytes(instanceIDBytes)
	sr.AttributeDescriptor.AttributeID = attributeID

	encodedValue, ok := decodedStruct[3].([]byte)
	if !ok {
		return fmt.Errorf("invalid type for Value")
	}
	value, err := axdr.Decode(encodedValue)
	if err != nil {
		return fmt.Errorf("failed to decode value: %w", err)
	}
	sr.Value = value

	return nil
}

// Encode encodes the ActionRequest APDU into a byte slice.
func (ar *ActionRequest) Encode() ([]byte, error) {
	obisBytes := ar.MethodDescriptor.InstanceID.Bytes()
	methodDescStruct := axdr.Structure{
		ar.MethodDescriptor.ClassID,
		obisBytes[:],
		ar.MethodDescriptor.MethodID,
	}

	encodedParams, err := axdr.Encode(ar.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode parameters: %w", err)
	}

	reqStruct := axdr.Structure{
		byte(ar.Type),
		ar.InvokeIDAndPriority,
		methodDescStruct,
		encodedParams,
	}

	reqBody, err := axdr.Encode(reqStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to encode ActionRequest body: %w", err)
	}

	res := append([]byte{byte(APDU_ACTION_REQUEST)}, reqBody...)
	return res, nil
}

// Decode decodes a byte slice into an ActionRequest APDU.
func (ar *ActionRequest) Decode(src []byte) error {
	if len(src) == 0 {
		return fmt.Errorf("empty source byte slice")
	}

	if APDUType(src[0]) != APDU_ACTION_REQUEST {
		return fmt.Errorf("invalid APDU tag for ActionRequest: got %X, expected %X", src[0], APDU_ACTION_REQUEST)
	}

	decoded, err := axdr.Decode(src[1:])
	if err != nil {
		return err
	}
	decodedStruct, ok := decoded.(axdr.Structure)
	if !ok {
		return fmt.Errorf("expected a structure")
	}

	if len(decodedStruct) != 4 {
		return fmt.Errorf("expected 4 elements in ActionRequest, got %d", len(decodedStruct))
	}

	reqType, ok := decodedStruct[0].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for ActionRequestType")
	}
	ar.Type = ActionRequestType(reqType)

	ar.InvokeIDAndPriority, ok = decodedStruct[1].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for InvokeIDAndPriority")
	}

	methodDescStruct, ok := decodedStruct[2].(axdr.Structure)
	if !ok {
		return fmt.Errorf("invalid type for MethodDescriptor")
	}

	if len(methodDescStruct) != 3 {
		return fmt.Errorf("expected 3 elements in MethodDescriptor, got %d", len(methodDescStruct))
	}

	classID, ok := methodDescStruct[0].(uint16)
	if !ok {
		return fmt.Errorf("invalid type for ClassID")
	}

	instanceID, ok := methodDescStruct[1].([]byte)
	if !ok {
		return fmt.Errorf("invalid type for InstanceID")
	}
	if len(instanceID) != 6 {
		return fmt.Errorf("invalid length for InstanceID: expected 6, got %d", len(instanceID))
	}

	methodID, ok := methodDescStruct[2].(int8)
	if !ok {
		return fmt.Errorf("invalid type for MethodID")
	}

	ar.MethodDescriptor.ClassID = classID
	var instanceIDBytes [6]byte
	copy(instanceIDBytes[:], instanceID)
	ar.MethodDescriptor.InstanceID.SetFromBytes(instanceIDBytes)
	ar.MethodDescriptor.MethodID = methodID

	encodedParams, ok := decodedStruct[3].([]byte)
	if !ok {
		return fmt.Errorf("invalid type for Parameters")
	}
	params, err := axdr.Decode(encodedParams)
	if err != nil {
		return fmt.Errorf("failed to decode parameters: %w", err)
	}
	ar.Parameters = params

	return nil
}

// Encode encodes the GetResponse APDU into a byte slice.
func (gr *GetResponse) Encode() ([]byte, error) {
	var resultChoice Choice
	if gr.Result.IsDataAccessResult {
		resultChoice = Choice{
			Tag:   1, // tag for data-access-result
			Value: byte(gr.Result.Value.(DataAccessResultEnum)),
		}
	} else {
		resultChoice = Choice{
			Tag:   0, // tag for data
			Value: gr.Result.Value,
		}
	}

	encodedChoice, err := axdr.Encode(resultChoice.Value)
	if err != nil {
		return nil, err
	}
	encodedChoice = append([]byte{resultChoice.Tag}, encodedChoice...)

	respStruct := axdr.Structure{
		byte(gr.Type),
		gr.InvokeIDAndPriority,
		encodedChoice,
	}

	respBody, err := axdr.Encode(respStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to encode GetResponse body: %w", err)
	}

	res := append([]byte{byte(APDU_GET_RESPONSE)}, respBody...)
	return res, nil
}

// Decode decodes a byte slice into a GetResponse APDU.
func (gr *GetResponse) Decode(src []byte) error {
	if len(src) == 0 {
		return fmt.Errorf("empty source byte slice")
	}

	if APDUType(src[0]) != APDU_GET_RESPONSE {
		return fmt.Errorf("invalid APDU tag for GetResponse: got %X, expected %X", src[0], APDU_GET_RESPONSE)
	}

	decoded, err := axdr.Decode(src[1:])
	if err != nil {
		return err
	}
	decodedStruct, ok := decoded.(axdr.Structure)
	if !ok {
		return fmt.Errorf("expected a structure")
	}

	if len(decodedStruct) != 3 {
		return fmt.Errorf("expected 3 elements in GetResponse, got %d", len(decodedStruct))
	}

	respType, ok := decodedStruct[0].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for GetResponseType")
	}
	gr.Type = GetResponseType(respType)

	gr.InvokeIDAndPriority, ok = decodedStruct[1].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for InvokeIDAndPriority")
	}

	encodedChoice, ok := decodedStruct[2].([]byte)
	if !ok {
		return fmt.Errorf("invalid type for GetDataResult")
	}
	if len(encodedChoice) < 1 {
		return fmt.Errorf("invalid length for GetDataResult")
	}
	resultTag := encodedChoice[0]
	resultValue, err := axdr.Decode(encodedChoice[1:])
	if err != nil {
		return fmt.Errorf("failed to decode GetDataResult value: %w", err)
	}

	switch resultTag {
	case 0: // data
		gr.Result.IsDataAccessResult = false
		gr.Result.Value = resultValue
	case 1: // data-access-result
		gr.Result.IsDataAccessResult = true
		val, ok := resultValue.(uint8)
		if !ok {
			return fmt.Errorf("invalid type for DataAccessResult")
		}
		gr.Result.Value = DataAccessResultEnum(val)
	default:
		return fmt.Errorf("invalid tag for GetDataResult CHOICE: %d", resultTag)
	}

	return nil
}

// Encode encodes the SetResponse APDU into a byte slice.
func (sr *SetResponse) Encode() ([]byte, error) {
	respStruct := axdr.Structure{
		byte(sr.Type),
		sr.InvokeIDAndPriority,
		byte(sr.Result),
	}

	respBody, err := axdr.Encode(respStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to encode SetResponse body: %w", err)
	}

	res := append([]byte{byte(APDU_SET_RESPONSE)}, respBody...)
	return res, nil
}

// Decode decodes a byte slice into a SetResponse APDU.
func (sr *SetResponse) Decode(src []byte) error {
	if len(src) == 0 {
		return fmt.Errorf("empty source byte slice")
	}

	if APDUType(src[0]) != APDU_SET_RESPONSE {
		return fmt.Errorf("invalid APDU tag for SetResponse: got %X, expected %X", src[0], APDU_SET_RESPONSE)
	}

	decoded, err := axdr.Decode(src[1:])
	if err != nil {
		return err
	}
	decodedStruct, ok := decoded.(axdr.Structure)
	if !ok {
		return fmt.Errorf("expected a structure")
	}

	if len(decodedStruct) != 3 {
		return fmt.Errorf("expected 3 elements in SetResponse, got %d", len(decodedStruct))
	}

	respType, ok := decodedStruct[0].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for SetResponseType")
	}
	sr.Type = SetResponseType(respType)

	sr.InvokeIDAndPriority, ok = decodedStruct[1].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for InvokeIDAndPriority")
	}

	result, ok := decodedStruct[2].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for Result")
	}
	sr.Result = DataAccessResultEnum(result)

	return nil
}

// Encode encodes the ActionResponse APDU into a byte slice.
func (ar *ActionResponse) Encode() ([]byte, error) {
	var resultChoice Choice
	if ar.Result.IsDataAccessResult {
		resultChoice = Choice{
			Tag:   1, // tag for data-access-result
			Value: byte(ar.Result.Value.(DataAccessResultEnum)),
		}
	} else {
		resultChoice = Choice{
			Tag:   0, // tag for data
			Value: ar.Result.Value,
		}
	}

	encodedChoice, err := axdr.Encode(resultChoice.Value)
	if err != nil {
		return nil, err
	}
	encodedChoice = append([]byte{resultChoice.Tag}, encodedChoice...)

	respStruct := axdr.Structure{
		byte(ar.Type),
		ar.InvokeIDAndPriority,
		encodedChoice,
	}

	respBody, err := axdr.Encode(respStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to encode ActionResponse body: %w", err)
	}

	res := append([]byte{byte(APDU_ACTION_RESPONSE)}, respBody...)
	return res, nil
}

// Decode decodes a byte slice into an ActionResponse APDU.
func (ar *ActionResponse) Decode(src []byte) error {
	if len(src) == 0 {
		return fmt.Errorf("empty source byte slice")
	}

	if APDUType(src[0]) != APDU_ACTION_RESPONSE {
		return fmt.Errorf("invalid APDU tag for ActionResponse: got %X, expected %X", src[0], APDU_ACTION_RESPONSE)
	}

	decoded, err := axdr.Decode(src[1:])
	if err != nil {
		return err
	}
	decodedStruct, ok := decoded.(axdr.Structure)
	if !ok {
		return fmt.Errorf("expected a structure")
	}

	if len(decodedStruct) != 3 {
		return fmt.Errorf("expected 3 elements in ActionResponse, got %d", len(decodedStruct))
	}

	respType, ok := decodedStruct[0].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for ActionResponseType")
	}
	ar.Type = ActionResponseType(respType)

	ar.InvokeIDAndPriority, ok = decodedStruct[1].(uint8)
	if !ok {
		return fmt.Errorf("invalid type for InvokeIDAndPriority")
	}

	encodedChoice, ok := decodedStruct[2].([]byte)
	if !ok {
		return fmt.Errorf("invalid type for ActionResult")
	}
	if len(encodedChoice) < 1 {
		return fmt.Errorf("invalid length for ActionResult")
	}
	resultTag := encodedChoice[0]
	resultValue, err := axdr.Decode(encodedChoice[1:])
	if err != nil {
		return fmt.Errorf("failed to decode ActionResult value: %w", err)
	}

	switch resultTag {
	case 0: // data
		ar.Result.IsDataAccessResult = false
		ar.Result.Value = resultValue
	case 1: // data-access-result
		ar.Result.IsDataAccessResult = true
		val, ok := resultValue.(uint8)
		if !ok {
			return fmt.Errorf("invalid type for DataAccessResult")
		}
		ar.Result.Value = DataAccessResultEnum(val)
	default:
		return fmt.Errorf("invalid tag for ActionResult CHOICE: %d", resultTag)
	}

	return nil
}
