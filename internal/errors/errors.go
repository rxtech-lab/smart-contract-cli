package errors

import (
	"errors"
)

// NewWithDetails creates a new custom error with additional details.
func NewWithDetails(code ErrorCode, message, details string) *CustomError {
	return &CustomError{
		code:    code,
		message: message,
		details: details,
	}
}

// Wrap wraps an existing error with a custom error code and message.
func Wrap(err error, code ErrorCode, message string) *CustomError {
	if err == nil {
		return nil
	}
	return &CustomError{
		code:    code,
		message: message,
		err:     err,
	}
}

// WrapWithDetails wraps an existing error with code, message, and details.
func WrapWithDetails(err error, code ErrorCode, message, details string) *CustomError {
	if err == nil {
		return nil
	}
	return &CustomError{
		code:    code,
		message: message,
		details: details,
		err:     err,
	}
}

// GetCode extracts the error code from an error if it's a CustomError.
func GetCode(err error) ErrorCode {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.Code()
	}
	return ErrCodeUnknown
}

// HasCode checks if an error has a specific error code.
func HasCode(err error, code ErrorCode) bool {
	return GetCode(err) == code
}

// ABI Domain Error Constructors

// NewABIError creates a new ABI-related error.
func NewABIError(code ErrorCode, message string) *CustomError {
	return New(code, message)
}

// WrapABIError wraps an error with an ABI error code.
func WrapABIError(err error, code ErrorCode, message string) *CustomError {
	return Wrap(err, code, message)
}

// Signer Domain Error Constructors

// NewSignerError creates a new signer-related error.
func NewSignerError(code ErrorCode, message string) *CustomError {
	return New(code, message)
}

// NewSignerErrorWithDetails creates a new signer error with details.
func NewSignerErrorWithDetails(code ErrorCode, message, details string) *CustomError {
	return NewWithDetails(code, message, details)
}

// WrapSignerError wraps an error with a signer error code.
func WrapSignerError(err error, code ErrorCode, message string) *CustomError {
	return Wrap(err, code, message)
}

// Transport Domain Error Constructors

// NewTransportError creates a new transport-related error.
func NewTransportError(code ErrorCode, message string) *CustomError {
	return New(code, message)
}

// WrapTransportError wraps an error with a transport error code.
func WrapTransportError(err error, code ErrorCode, message string) *CustomError {
	return Wrap(err, code, message)
}

// Contract Domain Error Constructors

// NewContractError creates a new contract-related error.
func NewContractError(code ErrorCode, message string) *CustomError {
	return New(code, message)
}

// WrapContractError wraps an error with a contract error code.
func WrapContractError(err error, code ErrorCode, message string) *CustomError {
	return Wrap(err, code, message)
}

// Database Domain Error Constructors

// NewDatabaseError creates a new database-related error.
func NewDatabaseError(code ErrorCode, message string) *CustomError {
	return New(code, message)
}

// WrapDatabaseError wraps an error with a database error code.
func WrapDatabaseError(err error, code ErrorCode, message string) *CustomError {
	return Wrap(err, code, message)
}
