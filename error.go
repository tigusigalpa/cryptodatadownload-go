package cryptodatadownload

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError is a non-2xx response from the CryptoDataDownload API.
//
// The OpenAPI schema declares 200 everywhere, 400 on most operations and 404
// on partition reads, but 401 and 429 are part of the real HTTP contract even
// though they are not declared — branch on the Is*() predicates or on Status.
//
// On 404 the API means "partition not found" (no data for that date), which
// is a data state rather than a routing failure.
type APIError struct {
	// Status is the authoritative HTTP response status code.
	Status int
	// Message is the server-supplied detail with echoed API keys redacted.
	Message string
	// Op names the SDK operation that produced the error.
	Op string
	// Headers are the original response headers, useful for Retry-After.
	Headers http.Header
	// RawBody contains bounded response diagnostics with echoed keys redacted.
	RawBody []byte
	// StatusCode is the HTTP status, equivalent to Status.
	StatusCode int
	// Response contains sanitized request and response metadata.
	Response *Response
	// BodyTruncated reports incomplete or size-limited diagnostics.
	BodyTruncated bool
}

// AuthenticationError represents HTTP 401 and wraps the common API error.
type AuthenticationError struct{ *APIError }

// Unwrap exposes the common API error for errors.As.
func (e *AuthenticationError) Unwrap() error { return e.APIError }

// BadRequestError represents HTTP 400.
type BadRequestError struct{ *APIError }

// Unwrap exposes the common API error for errors.As.
func (e *BadRequestError) Unwrap() error { return e.APIError }

// NotFoundError represents HTTP 404, including missing date partitions.
type NotFoundError struct{ *APIError }

// Unwrap exposes the common API error for errors.As.
func (e *NotFoundError) Unwrap() error { return e.APIError }

// RateLimitError represents HTTP 429. RetryAfter on the embedded APIError
// parses the server's delta-seconds or HTTP-date recommendation.
type RateLimitError struct{ *APIError }

// Unwrap exposes the common API error for errors.As.
func (e *RateLimitError) Unwrap() error { return e.APIError }

// ServerError represents HTTP 5xx, including statuses that are not retryable.
type ServerError struct{ *APIError }

// Unwrap exposes the common API error for errors.As.
func (e *ServerError) Unwrap() error { return e.APIError }

func classifyAPIError(e *APIError) error {
	switch {
	case e.IsAuthentication():
		return &AuthenticationError{e}
	case e.IsValidation():
		return &BadRequestError{e}
	case e.IsNotFound():
		return &NotFoundError{e}
	case e.IsRateLimited():
		return &RateLimitError{e}
	case e.IsServerError():
		return &ServerError{e}
	default:
		return e
	}
}

// Error implements the error interface. The message never includes request
// credentials.
func (e *APIError) Error() string {
	if e == nil {
		return "cryptodatadownload: <nil api error>"
	}
	if e.Message != "" {
		return fmt.Sprintf("cryptodatadownload API error %s %d: %s", e.Op, e.Status, e.Message)
	}
	return fmt.Sprintf("cryptodatadownload API error %s %d", e.Op, e.Status)
}

// IsValidation reports a 400 validation failure.
func (e *APIError) IsValidation() bool { return e != nil && e.Status == http.StatusBadRequest }

// IsAuthentication reports a 401 authentication failure.
func (e *APIError) IsAuthentication() bool {
	return e != nil && e.Status == http.StatusUnauthorized
}

// IsPermission reports a 403 permission failure.
func (e *APIError) IsPermission() bool { return e != nil && e.Status == http.StatusForbidden }

// IsNotFound reports a 404 failure — on partition endpoints this means "no
// data for that date yet", a data state rather than a routing failure.
func (e *APIError) IsNotFound() bool { return e != nil && e.Status == http.StatusNotFound }

// IsRateLimited reports a 429 rate-limited failure. Inspect RetryAfter for
// the server-recommended wait.
func (e *APIError) IsRateLimited() bool {
	return e != nil && e.Status == http.StatusTooManyRequests
}

// IsServerError reports a 5xx server failure.
func (e *APIError) IsServerError() bool { return e != nil && e.Status >= 500 && e.Status < 600 }

// RetryAfter returns the server-recommended wait from the Retry-After
// header, or zero if absent or unparseable.
func (e *APIError) RetryAfter() (seconds int, ok bool) {
	if e == nil || e.Headers == nil {
		return 0, false
	}
	return parseRetryAfter(e.Headers.Get("Retry-After"))
}

// AsAPIError returns the *APIError if err is (or wraps) one, else nil.
func AsAPIError(err error) *APIError {
	var e *APIError
	if errors.As(err, &e) {
		return e
	}
	return nil
}

// TransportError describes a failure below the API contract: a network
// error, an oversized or malformed response body, or a cancelled context.
// It never includes the API key.
type TransportError struct {
	// Op is a short label for the failing operation.
	Op string
	// StatusCode is the HTTP status, or zero for a network failure.
	StatusCode int
	// Err is the underlying error.
	Err error
}

// DecodeError describes malformed JSON, an unexpected content type or an
// unexpected result shape. It also unwraps to TransportError for compatibility.
type DecodeError struct{ *TransportError }

// Unwrap exposes the compatibility TransportError and its underlying cause.
func (e *DecodeError) Unwrap() error { return e.TransportError }

// Error implements the error interface.
func (e *TransportError) Error() string {
	if e == nil {
		return "cryptodatadownload: <nil transport error>"
	}
	if e.Err != nil {
		return fmt.Sprintf("cryptodatadownload %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("cryptodatadownload %s: transport error", e.Op)
}

// Unwrap returns the underlying error.
func (e *TransportError) Unwrap() error { return e.Err }

// AsTransport returns the *TransportError if err is (or wraps) one, else nil.
func AsTransport(err error) *TransportError {
	var t *TransportError
	if errors.As(err, &t) {
		return t
	}
	return nil
}

// ArgumentError is a client-side validation failure raised before any network
// I/O: a missing required parameter, an out-of-enum value or a bad date.
type ArgumentError struct {
	// Param is the offending parameter's wire name.
	Param string
	// Msg describes the violation.
	Msg string
}

// Error implements the error interface.
func (e *ArgumentError) Error() string {
	return fmt.Sprintf("cryptodatadownload: invalid %s: %s", e.Param, e.Msg)
}

// AsArgument returns the *ArgumentError if err is (or wraps) one, else nil.
func AsArgument(err error) *ArgumentError {
	var a *ArgumentError
	if errors.As(err, &a) {
		return a
	}
	return nil
}
