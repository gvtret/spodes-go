package cosem

import (
	"fmt"
	"github.com/gvtret/spodes-go/pkg/axdr"
)

// Application represents a COSEM application layer instance.
// It holds and manages all the COSEM objects.
type Application struct {
	objects          map[string]BaseInterface
	associationLN    *AssociationLN
	securitySetup    *SecuritySetup
	lastFrameCounter uint32
}

// NewApplication creates a new COSEM application instance.
func NewApplication(associationLN *AssociationLN, securitySetup *SecuritySetup) *Application {
	app := &Application{
		objects:       make(map[string]BaseInterface),
		associationLN: associationLN,
		securitySetup: securitySetup,
	}
	// Register the AssociationLN and SecuritySetup objects themselves
	app.RegisterObject(associationLN)
	app.RegisterObject(securitySetup)
	return app
}

// RegisterObject adds a COSEM object to the application.
// If an object with the same instance ID already exists, it will be overwritten.
func (app *Application) RegisterObject(obj BaseInterface) {
	instanceID := obj.GetInstanceID().String()
	app.objects[instanceID] = obj

	// Add the object to the AssociationLN's object list
	if app.associationLN != nil && obj.GetClassID() != AssociationLNClassID {
		app.associationLN.AddObject(obj)
	}
}

// FindObject retrieves a COSEM object by its instance ID (OBIS code).
// It returns the object and true if found, otherwise nil and false.
func (app *Application) FindObject(instanceID ObisCode) (BaseInterface, bool) {
	obj, found := app.objects[instanceID.String()]
	return obj, found
}

// HandleAPDU processes an incoming APDU.
func (app *Application) HandleAPDU(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("empty APDU")
	}

	apduType := APDUType(src[0])

	policy, err := app.securitySetup.GetAttribute(2)
	if err != nil {
		return nil, err
	}
	securityPolicy := policy.(SecurityPolicy)

	// Handle secured APDUs
	switch apduType {
	case APDU_GLO_GET_REQUEST, APDU_GLO_SET_REQUEST, APDU_GLO_ACTION_REQUEST:
		header := &SecurityHeader{}
		err := header.Decode(src[1:])
		if err != nil {
			return nil, err
		}

		// Check security policy
		sc := header.SecurityControl
		if (securityPolicy&PolicyAuthenticatedRequest != 0) && (sc != SecurityControlAuthenticationOnly && sc != SecurityControlAuthenticatedAndEncrypted) {
			return nil, fmt.Errorf("security policy violation: authenticated request required")
		}
		if (securityPolicy&PolicyEncryptedRequest != 0) && (sc != SecurityControlEncryptionOnly && sc != SecurityControlAuthenticatedAndEncrypted) {
			return nil, fmt.Errorf("security policy violation: encrypted request required")
		}

		var key []byte
		suite, err := app.securitySetup.GetAttribute(3)
		if err != nil {
			return nil, err
		}

		if sc == SecurityControlAuthenticatedAndEncrypted || sc == SecurityControlEncryptionOnly {
			key = app.securitySetup.GlobalUnicastKey
		} else {
			key = app.securitySetup.GlobalAuthenticationKey
		}

		serverSystemTitle, err := app.securitySetup.GetAttribute(5)
		if err != nil {
			return nil, err
		}

		plaintext, err := DecryptAndVerify(key, src[6:], serverSystemTitle.([]byte), header, suite.(SecuritySuite), app.lastFrameCounter)
		if err != nil {
			return nil, err
		}
		app.lastFrameCounter = header.FrameCounter

		// Dispatch to the appropriate handler
		var respAPDU APDU
		switch APDUType(plaintext[0]) {
		case APDU_GET_REQUEST:
			req := &GetRequest{}
			err = req.Decode(plaintext)
			if err != nil {
				return nil, err
			}
			respAPDU = app.HandleGetRequest(req)
		case APDU_SET_REQUEST:
			req := &SetRequest{}
			err = req.Decode(plaintext)
			if err != nil {
				return nil, err
			}
			respAPDU = app.HandleSetRequest(req)
		case APDU_ACTION_REQUEST:
			req := &ActionRequest{}
			err = req.Decode(plaintext)
			if err != nil {
				return nil, err
			}
			respAPDU = app.HandleActionRequest(req)
		default:
			return nil, fmt.Errorf("unsupported APDU type in secured APDU: %X", plaintext[0])
		}

		// Wrap the response in a secured APDU
		encodedResp, err := respAPDU.Encode()
		if err != nil {
			return nil, err
		}

		respHeader := &SecurityHeader{
			SecurityControl: header.SecurityControl,
			FrameCounter:    header.FrameCounter,
		}

		ciphertext, err := EncryptAndTag(key, encodedResp, serverSystemTitle.([]byte), respHeader, suite.(SecuritySuite))
		if err != nil {
			return nil, err
		}

		encodedRespHeader, err := respHeader.Encode()
		if err != nil {
			return nil, err
		}

		// Determine the response APDU type
		var respAPDUType APDUType
		switch apduType {
		case APDU_GLO_GET_REQUEST:
			respAPDUType = APDU_GLO_GET_RESPONSE
		case APDU_GLO_SET_REQUEST:
			respAPDUType = APDU_GLO_SET_RESPONSE
		case APDU_GLO_ACTION_REQUEST:
			respAPDUType = APDU_GLO_ACTION_RESPONSE
		}

		return append([]byte{byte(respAPDUType)}, append(encodedRespHeader, ciphertext...)...), nil

	// Handle unsecured APDUs
	case APDU_GET_REQUEST, APDU_SET_REQUEST, APDU_ACTION_REQUEST:
		if securityPolicy != PolicyNone {
			return nil, fmt.Errorf("security policy violation: unsecured request not allowed")
		}

		switch apduType {
		case APDU_GET_REQUEST:
			req := &GetRequest{}
			err := req.Decode(src)
			if err != nil {
				return nil, err
			}
			return app.HandleGetRequest(req).Encode()
		case APDU_SET_REQUEST:
			req := &SetRequest{}
			err := req.Decode(src)
			if err != nil {
				return nil, err
			}
			return app.HandleSetRequest(req).Encode()
		case APDU_ACTION_REQUEST:
			req := &ActionRequest{}
			err := req.Decode(src)
			if err != nil {
				return nil, err
			}
			return app.HandleActionRequest(req).Encode()
		}
	}
	return nil, fmt.Errorf("unsupported APDU type: %X", apduType)
}

// APDU is an interface for all APDU types.
type APDU interface {
	Encode() ([]byte, error)
}

// HandleGetRequest processes a Get-Request APDU and returns a Get-Response APDU.
func (app *Application) HandleGetRequest(req *GetRequest) *GetResponse {
	resp := &GetResponse{
		Type:                GET_RESPONSE_NORMAL,
		InvokeIDAndPriority: req.InvokeIDAndPriority,
	}

	obj, found := app.FindObject(req.AttributeDescriptor.InstanceID)
	if !found {
		resp.Result = GetDataResult{
			IsDataAccessResult: true,
			Value:              OBJECT_UNDEFINED,
		}
		return resp
	}

	val, err := obj.GetAttribute(byte(req.AttributeDescriptor.AttributeID))
	if err != nil {
		switch err {
		case ErrAttributeNotSupported:
			resp.Result = GetDataResult{
				IsDataAccessResult: true,
				Value:              OBJECT_UNAVAILABLE,
			}
		case ErrAccessDenied:
			resp.Result = GetDataResult{
				IsDataAccessResult: true,
				Value:              READ_WRITE_DENIED,
			}
		default:
			resp.Result = GetDataResult{
				IsDataAccessResult: true,
				Value:              OTHER_REASON,
			}
		}
		return resp
	}

	resp.Result = GetDataResult{
		IsDataAccessResult: false,
		Value:              val,
	}
	return resp
}

// HandleSetRequest processes a Set-Request APDU and returns a Set-Response APDU.
func (app *Application) HandleSetRequest(req *SetRequest) *SetResponse {
	resp := &SetResponse{
		Type:                SET_RESPONSE_NORMAL,
		InvokeIDAndPriority: req.InvokeIDAndPriority,
	}

	obj, found := app.FindObject(req.AttributeDescriptor.InstanceID)
	if !found {
		resp.Result = OBJECT_UNDEFINED
		return resp
	}

	err := obj.SetAttribute(byte(req.AttributeDescriptor.AttributeID), req.Value)
	if err != nil {
		switch err {
		case ErrAttributeNotSupported:
			resp.Result = OBJECT_UNAVAILABLE
		case ErrAccessDenied:
			resp.Result = READ_WRITE_DENIED
		case ErrInvalidValueType:
			resp.Result = TYPE_UNMATCHED
		default:
			resp.Result = OTHER_REASON
		}
		return resp
	}

	resp.Result = SUCCESS
	return resp
}

// HandleActionRequest processes an Action-Request APDU and returns an Action-Response APDU.
func (app *Application) HandleActionRequest(req *ActionRequest) *ActionResponse {
	resp := &ActionResponse{
		Type:                ACTION_RESPONSE_NORMAL,
		InvokeIDAndPriority: req.InvokeIDAndPriority,
	}

	obj, found := app.FindObject(req.MethodDescriptor.InstanceID)
	if !found {
		resp.Result = ActionResult{
			IsDataAccessResult: true,
			Value:              OBJECT_UNDEFINED,
		}
		return resp
	}

	params, ok := req.Parameters.(axdr.Array)
	if !ok {
		// If parameters are nil (e.g. for a method with no parameters), the axdr decoder might return nil
		// In that case, we can treat it as an empty array.
		if req.Parameters == nil {
			params = axdr.Array{}
		} else {
			resp.Result = ActionResult{
				IsDataAccessResult: true,
				Value:              TYPE_UNMATCHED,
			}
			return resp
		}
	}

	val, err := obj.Invoke(byte(req.MethodDescriptor.MethodID), params)
	if err != nil {
		switch err {
		case ErrMethodNotSupported:
			resp.Result = ActionResult{
				IsDataAccessResult: true,
				Value:              OBJECT_UNAVAILABLE,
			}
		case ErrAccessDenied:
			resp.Result = ActionResult{
				IsDataAccessResult: true,
				Value:              READ_WRITE_DENIED,
			}
		case ErrInvalidParameter:
			resp.Result = ActionResult{
				IsDataAccessResult: true,
				Value:              TYPE_UNMATCHED,
			}
		default:
			resp.Result = ActionResult{
				IsDataAccessResult: true,
				Value:              OTHER_REASON,
			}
		}
		return resp
	}

	resp.Result = ActionResult{
		IsDataAccessResult: false,
		Value:              val,
	}
	return resp
}
