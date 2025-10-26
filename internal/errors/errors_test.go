package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ErrorsTestSuite struct {
	suite.Suite
}

func TestErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}

func (s *ErrorsTestSuite) TestNew() {
	s.Run("creates error with code and message", func() {
		err := New(ErrCodeInvalidABIFormat, "test message")

		s.NotNil(err)
		s.Equal(ErrCodeInvalidABIFormat, err.Code())
		s.Contains(err.Error(), "INVALID_ABI_FORMAT")
		s.Contains(err.Error(), "test message")
	})
}

func (s *ErrorsTestSuite) TestNewWithDetails() {
	s.Run("creates error with code, message, and details", func() {
		err := NewWithDetails(ErrCodeInvalidSignatureLength, "invalid length", "expected 65, got 32")

		s.NotNil(err)
		s.Equal(ErrCodeInvalidSignatureLength, err.Code())
		s.Contains(err.Error(), "INVALID_SIGNATURE_LENGTH")
		s.Contains(err.Error(), "invalid length")
		s.Contains(err.Error(), "expected 65, got 32")
	})
}

func (s *ErrorsTestSuite) TestWrap() {
	s.Run("wraps existing error", func() {
		originalErr := errors.New("original error")
		wrappedErr := Wrap(originalErr, ErrCodeConnectionFailed, "connection failed")

		s.NotNil(wrappedErr)
		s.Equal(ErrCodeConnectionFailed, wrappedErr.Code())
		s.Contains(wrappedErr.Error(), "CONNECTION_FAILED")
		s.Contains(wrappedErr.Error(), "connection failed")
		s.Contains(wrappedErr.Error(), "original error")

		// Test Unwrap
		s.Equal(originalErr, wrappedErr.Unwrap())
	})

	s.Run("returns nil when wrapping nil error", func() {
		wrappedErr := Wrap(nil, ErrCodeConnectionFailed, "connection failed")
		s.Nil(wrappedErr)
	})
}

func (s *ErrorsTestSuite) TestWrapWithDetails() {
	s.Run("wraps error with details", func() {
		originalErr := errors.New("json parse error")
		wrappedErr := WrapWithDetails(originalErr, ErrCodeInvalidABIFormat, "failed to parse", "invalid JSON syntax")

		s.NotNil(wrappedErr)
		s.Equal(ErrCodeInvalidABIFormat, wrappedErr.Code())
		s.Contains(wrappedErr.Error(), "INVALID_ABI_FORMAT")
		s.Contains(wrappedErr.Error(), "failed to parse")
		s.Contains(wrappedErr.Error(), "invalid JSON syntax")
		s.Contains(wrappedErr.Error(), "json parse error")
	})

	s.Run("returns nil when wrapping nil error", func() {
		wrappedErr := WrapWithDetails(nil, ErrCodeConnectionFailed, "message", "details")
		s.Nil(wrappedErr)
	})
}

func (s *ErrorsTestSuite) TestGetCode() {
	s.Run("extracts code from custom error", func() {
		err := New(ErrCodeInvalidPrivateKey, "invalid key")
		code := GetCode(err)

		s.Equal(ErrCodeInvalidPrivateKey, code)
	})

	s.Run("returns unknown for non-custom error", func() {
		err := errors.New("standard error")
		code := GetCode(err)

		s.Equal(ErrCodeUnknown, code)
	})

	s.Run("extracts code from wrapped custom error", func() {
		originalErr := errors.New("original")
		wrappedErr := Wrap(originalErr, ErrCodeSigningFailed, "signing failed")
		code := GetCode(wrappedErr)

		s.Equal(ErrCodeSigningFailed, code)
	})
}

func (s *ErrorsTestSuite) TestHasCode() {
	s.Run("returns true for matching code", func() {
		err := New(ErrCodeTransactionTimeout, "timeout")

		s.True(HasCode(err, ErrCodeTransactionTimeout))
	})

	s.Run("returns false for non-matching code", func() {
		err := New(ErrCodeTransactionTimeout, "timeout")

		s.False(HasCode(err, ErrCodeConnectionFailed))
	})

	s.Run("returns false for standard error", func() {
		err := errors.New("standard error")

		s.False(HasCode(err, ErrCodeTransactionTimeout))
	})
}

func (s *ErrorsTestSuite) TestIsMethod() {
	s.Run("matches same error code", func() {
		err1 := New(ErrCodeInvalidABIFormat, "error 1")
		err2 := New(ErrCodeInvalidABIFormat, "error 2")

		s.True(errors.Is(err1, err2))
	})

	s.Run("does not match different error code", func() {
		err1 := New(ErrCodeInvalidABIFormat, "error 1")
		err2 := New(ErrCodeConnectionFailed, "error 2")

		s.False(errors.Is(err1, err2))
	})
}

func (s *ErrorsTestSuite) TestErrorFormatting() {
	s.Run("formats error with only message", func() {
		err := New(ErrCodeEndpointRequired, "endpoint required")
		expected := "[ENDPOINT_REQUIRED] endpoint required"

		s.Equal(expected, err.Error())
	})

	s.Run("formats error with message and details", func() {
		err := NewWithDetails(ErrCodeInvalidSignatureLength, "invalid length", "got 32 bytes")

		s.Contains(err.Error(), "[INVALID_SIGNATURE_LENGTH]")
		s.Contains(err.Error(), "invalid length")
		s.Contains(err.Error(), "got 32 bytes")
	})

	s.Run("formats error with message and wrapped error", func() {
		original := errors.New("network error")
		err := Wrap(original, ErrCodeConnectionFailed, "connection failed")

		s.Contains(err.Error(), "[CONNECTION_FAILED]")
		s.Contains(err.Error(), "connection failed")
		s.Contains(err.Error(), "network error")
	})

	s.Run("formats error with all fields", func() {
		original := errors.New("parse error")
		err := WrapWithDetails(original, ErrCodeInvalidABIFormat, "ABI parsing failed", "invalid JSON")

		s.Contains(err.Error(), "[INVALID_ABI_FORMAT]")
		s.Contains(err.Error(), "ABI parsing failed")
		s.Contains(err.Error(), "invalid JSON")
		s.Contains(err.Error(), "parse error")
	})
}

func (s *ErrorsTestSuite) TestDomainSpecificConstructors() {
	s.Run("ABI domain constructors", func() {
		err := NewABIError(ErrCodeInvalidABIFormat, "invalid format")
		s.Equal(ErrCodeInvalidABIFormat, err.Code())

		original := errors.New("json error")
		wrappedErr := WrapABIError(original, ErrCodeABIParseFailed, "parse failed")
		s.Equal(ErrCodeABIParseFailed, wrappedErr.Code())
		s.Equal(original, wrappedErr.Unwrap())
	})

	s.Run("Signer domain constructors", func() {
		err := NewSignerError(ErrCodeInvalidPrivateKey, "invalid key")
		s.Equal(ErrCodeInvalidPrivateKey, err.Code())

		errWithDetails := NewSignerErrorWithDetails(ErrCodeInvalidSignatureLength, "invalid length", "expected 65")
		s.Equal(ErrCodeInvalidSignatureLength, errWithDetails.Code())
		s.Contains(errWithDetails.Error(), "expected 65")

		original := errors.New("crypto error")
		wrappedErr := WrapSignerError(original, ErrCodeSigningFailed, "signing failed")
		s.Equal(ErrCodeSigningFailed, wrappedErr.Code())
		s.Equal(original, wrappedErr.Unwrap())
	})

	s.Run("Transport domain constructors", func() {
		err := NewTransportError(ErrCodeEndpointRequired, "endpoint required")
		s.Equal(ErrCodeEndpointRequired, err.Code())

		original := errors.New("network error")
		wrappedErr := WrapTransportError(original, ErrCodeConnectionFailed, "connection failed")
		s.Equal(ErrCodeConnectionFailed, wrappedErr.Code())
		s.Equal(original, wrappedErr.Unwrap())
	})
}

func (s *ErrorsTestSuite) TestErrorChaining() {
	s.Run("works with errors.Is", func() {
		original := errors.New("original error")
		wrapped := Wrap(original, ErrCodeConnectionFailed, "connection failed")

		s.True(errors.Is(wrapped, original))
	})

	s.Run("works with errors.As", func() {
		err := New(ErrCodeInvalidABIFormat, "invalid format")

		var customErr *CustomError
		s.True(errors.As(err, &customErr))
		s.Equal(ErrCodeInvalidABIFormat, customErr.Code())
	})

	s.Run("works with nested wrapping", func() {
		original := errors.New("original")
		wrapped1 := Wrap(original, ErrCodeConnectionFailed, "connection failed")
		wrapped2 := Wrap(wrapped1, ErrCodeTransactionSendFailed, "transaction send failed")

		s.True(errors.Is(wrapped2, wrapped1))
		s.True(errors.Is(wrapped2, original))
	})
}

func (s *ErrorsTestSuite) TestErrorCodeString() {
	s.Run("converts error code to string", func() {
		code := ErrCodeInvalidABIFormat
		s.Equal("INVALID_ABI_FORMAT", code.String())
	})
}
