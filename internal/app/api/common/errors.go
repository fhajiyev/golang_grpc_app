package common

import (
	"fmt"
	"net/http"

	"github.com/Buzzvil/buzzlib-go/core"
)

const (
	// GENERAL
	codeUnknown        = 0
	codeBadRequest     = 1
	codeValidation     = 2
	codeSessionExpired = 100
)

// Referral res codes
const (
	ReferralDenied      = 21
	ReferralInvalid     = 22
	ReferralInvalidCode = 23
	ReferralError       = 24

	NotFoundDevice = 41

	ReferralVerifyError = 51
)

// NewBindError create a binding error
func NewBindError(e error) *core.HttpError {
	return &core.HttpError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Invalid request parameters\n%s", e.Error()),
	}
}

// NewSessionError create a session error
func NewSessionError(e error) *core.HttpError {
	return &core.HttpError{
		Code:    http.StatusUnauthorized,
		Message: fmt.Sprintf("Your session has expired\n%s", e.Error()),
	}
}

// NewQueryKeyError create a query key error
func NewQueryKeyError(e error) *core.HttpError {
	return &core.HttpError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Invalid query key\n%s", e.Error()),
	}
}

// NewValidationErrorf creates a validation error with provided code and message
func NewValidationErrorf(format string, args ...interface{}) *core.HttpError {
	return &core.HttpError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewNotFoundErrorf creates a resource not found error with provided code and message
func NewNotFoundErrorf(format string, args ...interface{}) *core.HttpError {
	return &core.HttpError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewInternalServerError create a timeout error
func NewInternalServerError(e error) *core.HttpError {
	return &core.HttpError{
		Code:    http.StatusInternalServerError,
		Message: e.Error(),
	}
}
