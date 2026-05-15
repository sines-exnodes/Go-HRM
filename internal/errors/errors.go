package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// Stable code strings consumed by the FE. Keep in sync with the FE error map.
const (
	CodeNotFound     = "not_found"
	CodeBadRequest   = "bad_request"
	CodeConflict     = "conflict"
	CodeForbidden    = "forbidden"
	CodeUnauthorized = "unauthorized"
	CodeInternal     = "internal_error"
)

// AppError is the canonical error type raised by services. The error
// middleware converts these into the standard JSON envelope.
type AppError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	HTTP    int            `json:"-"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil AppError>"
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithDetails returns a shallow copy with details merged. Useful for adding
// per-call context without mutating package-level error instances.
func (e *AppError) WithDetails(details map[string]any) *AppError {
	if e == nil {
		return nil
	}
	merged := make(map[string]any, len(e.Details)+len(details))
	for k, v := range e.Details {
		merged[k] = v
	}
	for k, v := range details {
		merged[k] = v
	}
	cp := *e
	cp.Details = merged
	return &cp
}

// As is a tiny helper so callers can write
//     if ae, ok := apperrors.As(err); ok { ... }
// instead of importing the stdlib errors package just for this.
func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func ErrNotFound(resource string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		HTTP:    http.StatusNotFound,
	}
}

func ErrBadRequest(msg string) *AppError {
	return &AppError{
		Code:    CodeBadRequest,
		Message: msg,
		HTTP:    http.StatusBadRequest,
	}
}

func ErrConflict(msg string) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: msg,
		HTTP:    http.StatusConflict,
	}
}

func ErrForbidden(msg string) *AppError {
	return &AppError{
		Code:    CodeForbidden,
		Message: msg,
		HTTP:    http.StatusForbidden,
	}
}

func ErrUnauthorized(msg string) *AppError {
	return &AppError{
		Code:    CodeUnauthorized,
		Message: msg,
		HTTP:    http.StatusUnauthorized,
	}
}

func ErrInternal(msg string) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: msg,
		HTTP:    http.StatusInternalServerError,
	}
}
