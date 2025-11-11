package common

import "fmt"

// ErrorCode defines a type for specific error codes within the application.
type ErrorCode int

// Defines specific error codes for different layers and conditions.
const (
	// General Errors
	ErrUnknown ErrorCode = iota
	// Transport Errors
	ErrConnectionFailed
	ErrSendFailed
	ErrReceiveFailed
	ErrReadTimeout
	ErrConnectionClosed
	// HDLC Errors
	ErrHDLCInvalidFrame
	ErrHdlcFrameRejected
	ErrHDLCWindowFull
	// COSEM Errors
	ErrCosemAPDUUngarsable
	ErrCosemObjectUnavailable
	ErrCosemObjectAccessDenied
	ErrCosemTypeMismatch
	ErrCosemSecurityPolicyViolation
)

// SpodesError is a custom error type for the application.
type SpodesError struct {
	Code    ErrorCode
	Message string
	cause   error
}

// NewError creates a new SpodesError.
func NewError(code ErrorCode, message string) *SpodesError {
	return &SpodesError{Code: code, Message: message}
}

// WrapError creates a new SpodesError that wraps an existing error.
func WrapError(code ErrorCode, message string, cause error) *SpodesError {
	return &SpodesError{Code: code, Message: message, cause: cause}
}

// Error returns the error message.
func (e *SpodesError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

// Cause returns the underlying cause of the error.
func (e *SpodesError) Cause() error {
	return e.cause
}
