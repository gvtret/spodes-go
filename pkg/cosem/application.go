package cosem

// Application represents a COSEM application layer instance.
// It holds and manages all the COSEM objects.
type Application struct {
	objects       map[string]BaseInterface
	associationLN *AssociationLN
}

// NewApplication creates a new COSEM application instance.
func NewApplication(associationLN *AssociationLN) *Application {
	app := &Application{
		objects:       make(map[string]BaseInterface),
		associationLN: associationLN,
	}
	// Register the AssociationLN object itself
	app.RegisterObject(associationLN)
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

	params, ok := req.Parameters.([]interface{})
	if !ok {
		resp.Result = ActionResult{
			IsDataAccessResult: true,
			Value:              TYPE_UNMATCHED,
		}
		return resp
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
