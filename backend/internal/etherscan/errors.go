package etherscan

import "fmt"

// ErrorCode classifies provider failures for coverage reporting.
type ErrorCode string

const (
	CodeRateLimited   ErrorCode = "RATE_LIMITED"
	CodeInvalidAPIKey ErrorCode = "INVALID_API_KEY"
	CodeInvalidAddr   ErrorCode = "INVALID_ADDRESS"
	CodeTimeout       ErrorCode = "TIMEOUT"
	CodeHTTP          ErrorCode = "HTTP_ERROR"
	CodeAPI           ErrorCode = "API_ERROR"
	CodeUnsupported   ErrorCode = "UNSUPPORTED_RESPONSE"
	CodeCanceled      ErrorCode = "CANCELED"
	CodeNetwork       ErrorCode = "NETWORK_ERROR"
)

// APIError is a structured Etherscan client error.
type APIError struct {
	Code      ErrorCode
	Message   string
	Retryable bool
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("etherscan %s: %s", e.Code, e.Message)
}

func newAPIError(code ErrorCode, message string, retryable bool) *APIError {
	return &APIError{Code: code, Message: message, Retryable: retryable}
}
