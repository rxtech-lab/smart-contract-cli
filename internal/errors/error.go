package errors

import "fmt"

// CustomError represents a custom error with error code and context.
type CustomError struct {
	code    ErrorCode
	message string
	details string
	err     error
}

// Error implements the error interface.
func (e *CustomError) Error() string {
	if e.details != "" && e.err != nil {
		return fmt.Sprintf("[%s] %s: %s: %v", e.code, e.message, e.details, e.err)
	}
	if e.details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.code, e.message, e.details)
	}
	if e.err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code, e.message, e.err)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.message)
}

// Unwrap returns the wrapped error for error chain compatibility.
func (e *CustomError) Unwrap() error {
	return e.err
}

// Code returns the error code.
func (e *CustomError) Code() ErrorCode {
	return e.code
}

// Is supports errors.Is comparison.
func (e *CustomError) Is(target error) bool {
	t, ok := target.(*CustomError)
	if !ok {
		return false
	}
	return e.code == t.code
}

// New creates a new custom error.
func New(code ErrorCode, message string) *CustomError {
	return &CustomError{
		code:    code,
		message: message,
	}
}
