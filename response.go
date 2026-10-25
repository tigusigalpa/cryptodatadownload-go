package cryptodatadownload

import (
	"io"
	"net/http"
)

// Result contains an operation's decoded result, including responses whose
// wire representation omits the result envelope.
type Result[T any] struct {
	// Result is the decoded payload without the wire envelope.
	Result T `json:"result"`
}

// Response describes a completed HTTP attempt without retaining credentials.
// Headers are cloned so callers can inspect or modify them independently.
type Response struct {
	// StatusCode is the final HTTP status.
	StatusCode int
	// Headers are cloned response headers with echoed credentials redacted.
	Headers http.Header
	// URL is the request URL with credentials redacted.
	URL string
	// Method is the HTTP method, currently always GET.
	Method string
	// Operation is the SDK service and method label.
	Operation string
	// Attempts is the number of HTTP attempts, including the initial request.
	Attempts int
}

// RawResponse is a streaming download with response metadata. Callers must
// close Body, including after WriteTo. Reading is governed by the request
// context and the HTTP client's timeout; no success-body limit is imposed.
type RawResponse struct {
	*Response
	Body io.ReadCloser
}

// WriteTo copies the download to w without buffering the entire response.
// It leaves Body open so ownership stays with the caller.
func (r *RawResponse) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, r.Body)
}

// Close releases the download's response body.
func (r *RawResponse) Close() error { return r.Body.Close() }
