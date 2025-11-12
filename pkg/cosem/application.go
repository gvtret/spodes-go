package cosem

import (
	"fmt"
	"net"

	"github.com/gvtret/spodes-go/pkg/axdr"
	"github.com/gvtret/spodes-go/pkg/transport"
)

// Application represents a COSEM application layer instance.
// It holds and manages all the COSEM objects.
type Application struct {
	objects             map[string]BaseInterface
	associations        map[string]*AssociationLN
	securitySetup       *SecuritySetup
	transport           transport.Transport
	lastFrameCounters   map[*AssociationLN]uint32
	serverFrameCounters map[*AssociationLN]uint32
}

// NewApplication creates a new COSEM application instance.
func NewApplication(transport transport.Transport, securitySetup *SecuritySetup) *Application {
	app := &Application{
		objects:             make(map[string]BaseInterface),
		associations:        make(map[string]*AssociationLN),
		transport:           transport,
		securitySetup:       securitySetup,
		lastFrameCounters:   make(map[*AssociationLN]uint32),
		serverFrameCounters: make(map[*AssociationLN]uint32),
	}
	// Register the SecuritySetup object
	app.RegisterObject(securitySetup)
	return app
}

// AddAssociation maps a client address string to a specific AssociationLN instance.
func (app *Application) AddAssociation(address string, assoc *AssociationLN) {
	app.associations[address] = assoc
	app.lastFrameCounters[assoc] = 0
	app.serverFrameCounters[assoc] = assoc.ServerInvocationCounter()
	// Register the AssociationLN object itself
	app.RegisterObject(assoc)
}

// RegisterObject adds a COSEM object to the application's master object list.
// If an object with the same instance ID already exists, it will be overwritten.
func (app *Application) RegisterObject(obj BaseInterface) {
	instanceID := obj.GetInstanceID().String()
	app.objects[instanceID] = obj
}

// PopulateObjectList adds a curated list of objects to a specific AssociationLN's object list.
// This allows for creating filtered views for different clients (e.g., public vs. admin).
func (app *Application) PopulateObjectList(assoc *AssociationLN, objectOBISs []ObisCode) error {
	for _, obis := range objectOBISs {
		obj, found := app.FindObject(obis)
		if !found {
			return fmt.Errorf("object with OBIS code %s not found in master list", obis.String())
		}
		assoc.AddObject(obj)
	}
	return nil
}

// FindObject retrieves a COSEM object by its instance ID (OBIS code).
// It returns the object and true if found, otherwise nil and false.
func (app *Application) FindObject(instanceID ObisCode) (BaseInterface, bool) {
	obj, found := app.objects[instanceID.String()]
	return obj, found
}

// HandleAPDU processes an incoming APDU from a specific client address.
func (app *Application) HandleAPDU(src []byte, clientAddr net.Addr) ([]byte, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("empty APDU")
	}

	assoc, ok := app.associations[clientAddr.String()]
	if !ok {
		return nil, fmt.Errorf("no association found for client address: %s", clientAddr.String())
	}

	apduType := APDUType(src[0])

	switch apduType {
	case APDU_GLO_GET_REQUEST, APDU_GLO_SET_REQUEST, APDU_GLO_ACTION_REQUEST:
		return app.handleSecuredAPDU(apduType, src, assoc)
	case APDU_GET_REQUEST, APDU_SET_REQUEST, APDU_ACTION_REQUEST:
		return app.handleUnsecuredAPDU(apduType, src, assoc)
	default:
		return nil, fmt.Errorf("unsupported APDU type: %X", apduType)
	}
}

func (app *Application) handleSecuredAPDU(apduType APDUType, src []byte, assoc *AssociationLN) ([]byte, error) {
	policy, err := app.securitySetup.GetAttribute(2)
	if err != nil {
		return nil, err
	}
	securityPolicy := policy.(SecurityPolicy)

	header := &SecurityHeader{}
	err = header.Decode(src[1:])
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

	lastFrameCounter := app.lastFrameCounters[assoc]

	plaintext, err := DecryptAndVerify(key, src[6:], serverSystemTitle.([]byte), header, suite.(SecuritySuite), lastFrameCounter)
	if err != nil {
		return nil, err
	}
	app.lastFrameCounters[assoc] = header.FrameCounter

	respAPDU, err := app.dispatchAPDU(plaintext, assoc)
	if err != nil {
		return nil, err
	}

	// Wrap the response in a secured APDU
	encodedResp, err := respAPDU.Encode()
	if err != nil {
		return nil, err
	}

	nextFrameCounter := app.serverFrameCounters[assoc] + 1
	app.serverFrameCounters[assoc] = nextFrameCounter
	assoc.SetServerInvocationCounter(nextFrameCounter)

	respHeader := &SecurityHeader{
		SecurityControl: header.SecurityControl,
		FrameCounter:    nextFrameCounter,
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
}

func (app *Application) handleUnsecuredAPDU(apduType APDUType, src []byte, assoc *AssociationLN) ([]byte, error) {
	policy, err := app.securitySetup.GetAttribute(2)
	if err != nil {
		return nil, err
	}
	securityPolicy := policy.(SecurityPolicy)

	if securityPolicy != PolicyNone {
		return nil, fmt.Errorf("security policy violation: unsecured request not allowed")
	}

	respAPDU, err := app.dispatchAPDU(src, assoc)
	if err != nil {
		return nil, err
	}

	return respAPDU.Encode()
}

func (app *Application) dispatchAPDU(src []byte, assoc *AssociationLN) (APDU, error) {
	apduType := APDUType(src[0])
	switch apduType {
	case APDU_GET_REQUEST:
		req := &GetRequest{}
		err := req.Decode(src)
		if err != nil {
			return nil, err
		}
		return app.HandleGetRequest(req, assoc), nil
	case APDU_SET_REQUEST:
		req := &SetRequest{}
		err := req.Decode(src)
		if err != nil {
			return nil, err
		}
		return app.HandleSetRequest(req, assoc), nil
	case APDU_ACTION_REQUEST:
		req := &ActionRequest{}
		err := req.Decode(src)
		if err != nil {
			return nil, err
		}
		return app.HandleActionRequest(req, assoc), nil
	default:
		return nil, fmt.Errorf("unsupported APDU type: %X", apduType)
	}
}

// APDU is an interface for all APDU types.
type APDU interface {
	Encode() ([]byte, error)
}

// HandleGetRequest processes a Get-Request APDU and returns a Get-Response APDU.
func (app *Application) HandleGetRequest(req *GetRequest, assoc *AssociationLN) *GetResponse {
	resp := &GetResponse{
		Type:                GET_RESPONSE_NORMAL,
		InvokeIDAndPriority: req.InvokeIDAndPriority,
	}

	if !assoc.CheckAttributeAccess(req.AttributeDescriptor.InstanceID, byte(req.AttributeDescriptor.AttributeID), Read) {
		resp.Result = GetDataResult{
			IsDataAccessResult: true,
			Value:              READ_WRITE_DENIED,
		}
		return resp
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
		// The original error from GetAttribute might be too generic.
		// We can provide a more specific access control error if the association check failed.
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
func (app *Application) HandleSetRequest(req *SetRequest, assoc *AssociationLN) *SetResponse {
	resp := &SetResponse{
		Type:                SET_RESPONSE_NORMAL,
		InvokeIDAndPriority: req.InvokeIDAndPriority,
	}

	if !assoc.CheckAttributeAccess(req.AttributeDescriptor.InstanceID, byte(req.AttributeDescriptor.AttributeID), Write) {
		resp.Result = READ_WRITE_DENIED
		return resp
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
func (app *Application) HandleActionRequest(req *ActionRequest, assoc *AssociationLN) *ActionResponse {
	resp := &ActionResponse{
		Type:                ACTION_RESPONSE_NORMAL,
		InvokeIDAndPriority: req.InvokeIDAndPriority,
	}

	if !assoc.CheckMethodAccess(req.MethodDescriptor.InstanceID, byte(req.MethodDescriptor.MethodID)) {
		resp.Result = ActionResult{
			IsDataAccessResult: true,
			Value:              READ_WRITE_DENIED,
		}
		return resp
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
