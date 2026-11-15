// Package apierror defines structured API errors without leaking internals.
package apierror

import (
	"fmt"
	"net/http"
)

// Code is a stable machine-readable error identifier.
type Code string

const (
	CodeInvalidRequest   Code = "INVALID_REQUEST"
	CodeInvalidConfig    Code = "INVALID_CONFIG"
	CodeNotFound         Code = "NOT_FOUND"
	CodeNotReady         Code = "NOT_READY"
	CodeInternalError    Code = "INTERNAL_ERROR"
	CodeTimeout          Code = "REQUEST_TIMEOUT"
	CodeUnsupportedChain Code = "UNSUPPORTED_CHAIN"
)

// Error is a client-safe API error.
type Error struct {
	Code       Code                   `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Response is the JSON error envelope.
type Response struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the nested error object in responses.
type ErrorBody struct {
	Code    Code                   `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// New builds an API error.
func New(status int, code Code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Details:    map[string]interface{}{},
	}
}

// WithDetails attaches non-sensitive detail fields.
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	if e == nil {
		return nil
	}
	e.Details = details
	return e
}

// ToResponse maps to the public envelope.
func (e *Error) ToResponse() Response {
	if e == nil {
		return Response{
			Error: ErrorBody{
				Code:    CodeInternalError,
				Message: "An unexpected error occurred.",
			},
		}
	}
	return Response{
		Error: ErrorBody{
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
		},
	}
}

// Helpers for common cases.

func BadRequest(code Code, message string) *Error {
	return New(http.StatusBadRequest, code, message)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, CodeNotFound, message)
}

func NotReady(message string) *Error {
	return New(http.StatusServiceUnavailable, CodeNotReady, message)
}

func Internal(message string) *Error {
	return New(http.StatusInternalServerError, CodeInternalError, message)
}

func Timeout(message string) *Error {
	return New(http.StatusGatewayTimeout, CodeTimeout, message)
}
