package cosem

import (
	"fmt"
	"reflect"
)

// AttibuteAccess - describe access level for COSEM object attribute.
type AttributeAccess byte

const (
	// No access.
	AttributeNoAccess AttributeAccess = 0x00
	// The client is allowed only reading from the server.
	AttributeRead AttributeAccess = 0x01
	// The client is allowed only writing to the server.
	AttributeWrite AttributeAccess = 0x02
	// Request messages are authenticated.
	AttributeAuthenticatedRequest AttributeAccess = 0x04
	// Request messages are encrypted.
	AttributeEncryptedRequest AttributeAccess = 0x08
	// Request messages are digitally signed.
	AttributeDigitallySignedRequest AttributeAccess = 0x10
	// Response messages are authenticated.
	AttributeAuthenticatedResponse AttributeAccess = 0x20
	// Response messages are encrypted.
	AttributeEncryptedResponse AttributeAccess = 0x40
	// Response messages are digitally signed.
	AttributeDigitallySignedResponse AttributeAccess = 0x80
)

// AttributeDescriptor - define COSEM object attribute
type AttributeDescriptor struct {
	Type      reflect.Type
	Access    AttributeAccess
	Value     interface{}
	Validator func(value interface{}) error
}

// MethodAccess - describe access level for COSEM object method.
type MethodAccess byte

const (
	// Client can't use method.
	MethodNoAccess MethodAccess = 0x0
	// Access is allowed.
	MethodAccessAllowed MethodAccess = 0x1
	// Authenticated request.
	MethodAuthenticatedRequest MethodAccess = 0x4
	// Encrypted request.
	MethodEncryptedRequest MethodAccess = 0x8
	// Digitally signed request.
	MethodDigitallySignedRequest MethodAccess = 0x10
	// Authenticated response.
	MethodAuthenticatedResponse MethodAccess = 0x20
	// Encrypted response.
	MethodEncryptedResponse MethodAccess = 0x40
	// Digitally signed response.
	MethodDigitallySignedResponse MethodAccess = 0x80
)

// MethodDescriptor - define COSEM object method
type MethodDescriptor struct {
	Access     MethodAccess
	ParamTypes []reflect.Type
	ReturnType reflect.Type
	Handler    func(params []interface{}) (interface{}, error)
}

// Error types
var (
	ErrAttributeNotSupported = fmt.Errorf("attribute not supported")
	ErrMethodNotSupported    = fmt.Errorf("method not supported")
	ErrAccessDenied          = fmt.Errorf("access denied")
	ErrInvalidParameter      = fmt.Errorf("invalid parameter")
	ErrInvalidValueType      = fmt.Errorf("invalid value type")
)

// Callback types
type (
	// PreReadCallback is invoked before an attribute is read.
	// It receives the attribute ID and context, and returns an error to abort the read.
	PreReadCallback func(attributeID byte, ctx interface{}) error

	// PostReadCallback is invoked after an attribute is successfully read.
	// It receives the attribute ID, the read value, and context.
	PostReadCallback func(attributeID byte, value interface{}, ctx interface{})

	// PreWriteCallback is invoked before an attribute is written.
	// It receives the attribute ID, the value to be written, and context.
	// It can return an error to abort the write.
	PreWriteCallback func(attributeID byte, value interface{}, ctx interface{}) error

	// PostWriteCallback is invoked after an attribute is successfully written.
	// It receives the attribute ID, the written value, and context.
	PostWriteCallback func(attributeID byte, value interface{}, ctx interface{})

	// PreActionCallback is invoked before a method is executed.
	// It receives the method ID, parameters, and context.
	// It can return an error to abort the execution.
	PreActionCallback func(methodID byte, params []interface{}, ctx interface{}) error

	// PostActionCallback is invoked after a method is successfully executed.
	// It receives the method ID, parameters, the result, and context.
	PostActionCallback func(methodID byte, params []interface{}, result interface{}, ctx interface{})
)

// BaseInterface is an interface for a GXDLMS object.
//
// It contains methods for getting and setting attributes, invoking methods, and getting attribute and method access.
type BaseInterface interface {
	GetClassID() uint16
	GetInstanceID() ObisCode
	GetAttribute(attributeID byte) (interface{}, error)
	SetAttribute(attributeID byte, value interface{}) error
	Invoke(methodID byte, parameters []interface{}) (interface{}, error)
	GetAttributeAccess(attributeID byte) AttributeAccess
	GetMethodAccess(methodID byte) MethodAccess
	SetPreReadCallback(cb PreReadCallback)
	SetPostReadCallback(cb PostReadCallback)
	SetPreWriteCallback(cb PreWriteCallback)
	SetPostWriteCallback(cb PostWriteCallback)
	SetPreActionCallback(cb PreActionCallback)
	SetPostActionCallback(cb PostActionCallback)
	SetCallbackContext(ctx interface{})
}

// BaseImpl is a base implementation of a COSEM object.
//
// It contains the class ID, instance ID, attributes, and methods of the object.
type BaseImpl struct {
	ClassID            uint16
	InstanceID         ObisCode
	Attributes         map[byte]AttributeDescriptor
	Methods            map[byte]MethodDescriptor
	PreReadCallback    PreReadCallback
	PostReadCallback   PostReadCallback
	PreWriteCallback   PreWriteCallback
	PostWriteCallback  PostWriteCallback
	PreActionCallback  PreActionCallback
	PostActionCallback PostActionCallback
	CallbackContext    interface{}
}

// GetClassID gets COSEM object class ID
func (b *BaseImpl) GetClassID() uint16 {
	return b.ClassID
}

// GetInstanceID gets COSEM object OBIS code
func (b *BaseImpl) GetInstanceID() ObisCode {
	return b.InstanceID
}

// GetAttribute gets the value of an attribute.
//
// The attributeID is the ID of the attribute to be retrieved.
func (b *BaseImpl) GetAttribute(attributeID byte) (interface{}, error) {
	if b.PreReadCallback != nil {
		if err := b.PreReadCallback(attributeID, b.CallbackContext); err != nil {
			return nil, err
		}
	}

	attr, exists := b.Attributes[attributeID]
	if !exists {
		return nil, ErrAttributeNotSupported
	}
	// Allow access for either AttributeRead or AttributeWrite
	if attr.Access&AttributeRead == 0 {
		return nil, ErrAccessDenied
	}

	if b.PostReadCallback != nil {
		b.PostReadCallback(attributeID, attr.Value, b.CallbackContext)
	}
	return attr.Value, nil
}

// SetAttribute sets the value of an attribute.
//
// The attributeID is the ID of the attribute to be set.
// The value is the value to be set.
func (b *BaseImpl) SetAttribute(attributeID byte, value interface{}) error {
	if b.PreWriteCallback != nil {
		if err := b.PreWriteCallback(attributeID, value, b.CallbackContext); err != nil {
			return err
		}
	}

	attr, exists := b.Attributes[attributeID]
	if !exists {
		return ErrAttributeNotSupported
	}
	if attr.Access&AttributeWrite == 0 {
		return ErrAccessDenied
	}
	if reflect.TypeOf(value) != attr.Type {
		return ErrInvalidValueType
	}

	if attr.Validator != nil {
		if err := attr.Validator(value); err != nil {
			return err
		}
	}
	attr.Value = value
	b.Attributes[attributeID] = attr

	if b.PostWriteCallback != nil {
		b.PostWriteCallback(attributeID, value, b.CallbackContext)
	}
	return nil
}

// Invoke invokes a method with the given parameters.
//
// The methodID is the ID of the method to be invoked.
// The parameters are the parameters to be passed to the method.
func (b *BaseImpl) Invoke(methodID byte, parameters []interface{}) (interface{}, error) {
	if b.PreActionCallback != nil {
		if err := b.PreActionCallback(methodID, parameters, b.CallbackContext); err != nil {
			return nil, err
		}
	}

	method, exists := b.Methods[methodID]
	if !exists {
		return nil, ErrMethodNotSupported
	}

	if method.Access == MethodNoAccess {
		return nil, ErrAccessDenied
	}

	if len(parameters) != len(method.ParamTypes) {
		return nil, ErrInvalidParameter
	}

	for i, param := range parameters {
		if reflect.TypeOf(param) != method.ParamTypes[i] {
			return nil, ErrInvalidParameter
		}
	}

	result, err := method.Handler(parameters)
	if err != nil {
		return nil, err
	}

	if b.PostActionCallback != nil {
		b.PostActionCallback(methodID, parameters, result, b.CallbackContext)
	}
	return result, nil
}

// GetAttributeAccess gets the access level of an attribute.
//
// The attributeID is the ID of the attribute to be retrieved.
func (b *BaseImpl) GetAttributeAccess(attributeID byte) AttributeAccess {
	if attr, exists := b.Attributes[attributeID]; exists {
		return attr.Access
	}
	return AttributeNoAccess
}

// GetMethodAccess gets the access level of an method.
//
// The methodID is the ID of the method to be retrieved.
func (b *BaseImpl) GetMethodAccess(methodID byte) MethodAccess {
	if method, exists := b.Methods[methodID]; exists {
		return method.Access
	}
	return MethodNoAccess
}

// SetPreReadCallback sets the pre-read callback.
func (b *BaseImpl) SetPreReadCallback(cb PreReadCallback) {
	b.PreReadCallback = cb
}

// SetPostReadCallback sets the post-read callback.
func (b *BaseImpl) SetPostReadCallback(cb PostReadCallback) {
	b.PostReadCallback = cb
}

// SetPreWriteCallback sets the pre-write callback.
func (b *BaseImpl) SetPreWriteCallback(cb PreWriteCallback) {
	b.PreWriteCallback = cb
}

// SetPostWriteCallback sets the post-write callback.
func (b *BaseImpl) SetPostWriteCallback(cb PostWriteCallback) {
	b.PostWriteCallback = cb
}

// SetPreActionCallback sets the pre-action callback.
func (b *BaseImpl) SetPreActionCallback(cb PreActionCallback) {
	b.PreActionCallback = cb
}

// SetPostActionCallback sets the post-action callback.
func (b *BaseImpl) SetPostActionCallback(cb PostActionCallback) {
	b.PostActionCallback = cb
}

// SetCallbackContext sets the callback context.
func (b *BaseImpl) SetCallbackContext(ctx interface{}) {
	b.CallbackContext = ctx
}
