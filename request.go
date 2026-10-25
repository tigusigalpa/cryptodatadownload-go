package cryptodatadownload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
)

// apiResponse is a decoded REST response. The body is fully buffered; the
// underlying connection has already been closed when do returns.
type apiResponse struct {
	statusCode int
	header     http.Header
	body       []byte
	metadata   *Response
	stream     io.ReadCloser
}

// do performs a GET request against the CryptoDataDownload API, applying
// Token authentication, the client-side rate limiter, retry policy and error
// mapping. path is the operation path relative to the base URL (e.g.
// "/data/ohlc/binance/spot/"). Authentication defaults to a Token header;
// query-token mode is explicit, and diagnostic metadata redacts credentials.
func (c *Client) do(ctx context.Context, op, path string, query url.Values) (*apiResponse, error) {
	return c.doRequest(ctx, op, path, query, false)
}

func (c *Client) doStream(ctx context.Context, op, path string, query url.Values) (*RawResponse, error) {
	resp, err := c.doRequest(ctx, op, path, query, true)
	if err != nil {
		return nil, err
	}
	return &RawResponse{Response: resp.metadata, Body: resp.stream}, nil
}

func (c *Client) doRequest(ctx context.Context, op, path string, query url.Values, stream bool) (*apiResponse, error) {
	if ctx == nil {
		return nil, c.transportError(op, 0, errors.New("nil context"))
	}
	if err := ctx.Err(); err != nil {
		return nil, c.transportError(op, 0, err)
	}
	if c.apiKey == "" {
		return nil, ErrAPIKeyRequired
	}

	attempts := c.cfg.retry.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}

	fullURL := c.buildURL(path, query)

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := c.waitForRateLimit(ctx); err != nil {
			return nil, c.transportError(op, 0, err)
		}

		resp, sendErr := c.sendOnce(ctx, fullURL)
		if sendErr != nil {
			lastErr = c.transportError(op, 0, sendErr)
			if attempt < attempts && transientTransportError(sendErr) && ctx.Err() == nil {
				d := c.backoff(attempt)
				c.logRetry(attempt, attempts, d)
				if waitErr := c.wait(ctx, d); waitErr != nil {
					return nil, c.transportError(op, 0, waitErr)
				}
				continue
			}
			return nil, lastErr
		}

		metadata := &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			URL:        c.sanitizedURL(fullURL),
			Method:     http.MethodGet,
			Operation:  op,
			Attempts:   attempt,
		}
		for key, values := range metadata.Headers {
			for i, value := range values {
				values[i] = redactCredential(value, c.apiKey)
			}
			metadata.Headers[key] = values
		}
		result := &apiResponse{statusCode: resp.StatusCode, header: resp.Header.Clone(), metadata: metadata}
		success := resp.StatusCode >= 200 && resp.StatusCode < 300
		if stream && success {
			result.stream = &redactedBody{ReadCloser: resp.Body, secret: c.apiKey}
			return result, nil
		}
		limit := c.cfg.maxResponseBodyBytes
		if !success {
			limit = c.cfg.maxErrorBodyBytes
		}
		body, readErr := readResponseBody(resp.Body, limit)
		_ = resp.Body.Close()
		if readErr != nil && (success || ctx.Err() != nil) {
			lastErr = c.transportError(op, resp.StatusCode, readErr)
			var tooLarge *responseBodyLimitError
			retryable := resp.StatusCode >= 200 && resp.StatusCode < 300 || retryableStatus(resp.StatusCode)
			if attempt < attempts && retryable && !errors.As(readErr, &tooLarge) {
				d := c.backoffWithRetryAfter(attempt, resp.Header)
				c.logRetry(attempt, attempts, d)
				if waitErr := c.wait(ctx, d); waitErr != nil {
					return nil, c.transportError(op, 0, waitErr)
				}
				continue
			}
			return result, lastErr
		}

		if success {
			result.body = body
			return result, nil
		}

		apiErr := parseError(op, resp, body)
		apiErr.Message = redactCredential(apiErr.Message, c.apiKey)
		if int64(len(apiErr.Message)) > c.cfg.maxErrorBodyBytes {
			apiErr.Message = apiErr.Message[:c.cfg.maxErrorBodyBytes]
		}
		apiErr.RawBody = []byte(redactCredential(string(apiErr.RawBody), c.apiKey))
		apiErr.Headers = metadata.Headers.Clone()
		apiErr.Response = metadata
		apiErr.StatusCode = resp.StatusCode
		apiErr.BodyTruncated = readErr != nil
		if int64(len(apiErr.RawBody)) > c.cfg.maxErrorBodyBytes {
			apiErr.RawBody = apiErr.RawBody[:c.cfg.maxErrorBodyBytes]
			apiErr.BodyTruncated = true
		}
		lastErr = classifyAPIError(apiErr)

		if !retryableStatus(resp.StatusCode) || attempt >= attempts {
			return result, lastErr
		}
		d := c.backoffWithRetryAfter(attempt, resp.Header)
		c.logRetry(attempt, attempts, d)
		if waitErr := c.wait(ctx, d); waitErr != nil {
			return nil, c.transportError(op, 0, waitErr)
		}
	}
	return nil, lastErr
}

func (c *Client) waitForRateLimit(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.cfg.rateLimiter != nil {
		if err := c.cfg.rateLimiter.Wait(ctx); err != nil {
			return err
		}
		return ctx.Err()
	}
	return c.gate.wait(ctx)
}

func transientTransportError(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	for _, cause := range []error{syscall.ECONNRESET, syscall.ECONNABORTED, syscall.ECONNREFUSED, syscall.EPIPE} {
		if errors.Is(err, cause) {
			return true
		}
	}
	var networkError interface {
		Timeout() bool
		Temporary() bool
	}
	return errors.As(err, &networkError) && (networkError.Timeout() || networkError.Temporary())
}

// redactedError preserves errors.Is/As while keeping echoed credentials out
// of the printable error. The original error remains available through Unwrap.
type redactedError struct {
	err    error
	secret string
}

type redactedBody struct {
	io.ReadCloser
	secret string
}

func (r *redactedBody) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if err != nil && err != io.EOF {
		err = &redactedError{err: err, secret: r.secret}
	}
	return n, err
}

func (r *redactedBody) Close() error {
	if err := r.ReadCloser.Close(); err != nil {
		return &redactedError{err: err, secret: r.secret}
	}
	return nil
}

func (e *redactedError) Error() string {
	return redactCredential(e.err.Error(), e.secret)
}

func (e *redactedError) Unwrap() error { return e.err }

func redactCredential(value, secret string) string {
	if secret == "" {
		return value
	}
	return strings.NewReplacer(
		url.QueryEscape(secret), "[REDACTED]",
		url.PathEscape(secret), "[REDACTED]",
		secret, "[REDACTED]",
	).Replace(value)
}

func (c *Client) transportError(op string, status int, err error) *TransportError {
	if c.apiKey != "" && err != nil {
		err = &redactedError{err: err, secret: c.apiKey}
	}
	return &TransportError{Op: op, StatusCode: status, Err: err}
}

func (c *Client) decodeError(op string, resp *apiResponse, err error) error {
	if original := AsTransport(err); original != nil {
		err = original.Err
	}
	return &DecodeError{c.transportError(op, resp.statusCode, err)}
}

func validateJSONContentType(header http.Header) error {
	value := header.Get("Content-Type")
	if value == "" {
		return nil
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return fmt.Errorf("invalid JSON content type: %w", err)
	}
	if mediaType != "application/json" && !(strings.HasPrefix(mediaType, "application/") && strings.HasSuffix(mediaType, "+json")) {
		return fmt.Errorf("expected JSON content type, got %q", mediaType)
	}
	return nil
}

// sendOnce performs a single HTTP GET without retry logic.
func (c *Client) sendOnce(ctx context.Context, fullURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if !c.cfg.queryTokenAuth {
		req.Header.Set("Authorization", "Token "+c.apiKey)
	}
	req.Header.Set("User-Agent", c.cfg.userAgent)

	return c.cfg.httpClient.Do(req)
}

// buildURL joins the base URL with the operation path and encoded query.
func (c *Client) buildURL(path string, query url.Values) string {
	u, _ := url.Parse(c.cfg.baseURL) // validated by NewClient
	escapedPath := strings.TrimRight(u.EscapedPath(), "/") + path
	u.Path = strings.TrimRight(u.Path, "/") + path
	u.RawPath = escapedPath
	values := make(url.Values, len(query)+1)
	for key, value := range query {
		values[key] = append([]string(nil), value...)
	}
	if c.cfg.queryTokenAuth {
		values.Set("auth_token", c.apiKey)
	}
	u.RawQuery = values.Encode()
	return u.String()
}

func (c *Client) sanitizedURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	query := u.Query()
	if query.Has("auth_token") {
		query.Set("auth_token", "[REDACTED]")
	}
	u.RawQuery = query.Encode()
	return redactCredential(u.String(), c.apiKey)
}

// logRetry emits a redacted diagnostic line for a pending retry.
func (c *Client) logRetry(attempt, attempts int, d time.Duration) {
	if c.cfg.logger != nil {
		c.cfg.logger.Printf("retrying GET after %s (attempt %d/%d)", d, attempt+1, attempts)
	}
}

// wait sleeps for d respecting context cancellation via the injected sleeper.
func (c *Client) wait(ctx context.Context, d time.Duration) error {
	return c.cfg.sleeper.Sleep(ctx, d)
}

// backoff computes capped exponential backoff with full jitter for attempt n.
func (c *Client) backoff(attempt int) time.Duration {
	initial := c.cfg.retry.InitialBackoff
	if initial <= 0 {
		initial = 500 * time.Millisecond
	}
	max := c.cfg.retry.MaxBackoff
	if max <= 0 {
		max = 30 * time.Second
	}
	d := initial
	if d > max {
		d = max
	}
	for i := 1; i < attempt; i++ {
		if d >= max-d {
			d = max
			break
		}
		d *= 2
	}
	// Full jitter: randomize within [0, d].
	if d > 0 {
		d = time.Duration(rand.Int63n(int64(d)))
	}
	return d
}

// backoffWithRetryAfter is backoff that honours the Retry-After header when
// it requests a longer wait than the computed backoff.
func (c *Client) backoffWithRetryAfter(attempt int, header http.Header) time.Duration {
	d := c.backoff(attempt)
	if header != nil {
		if secs, ok := parseRetryAfterAt(header.Get("Retry-After"), c.cfg.clock.Now()); ok && secs > 0 {
			ra := time.Duration(math.MaxInt64)
			if int64(secs) <= math.MaxInt64/int64(time.Second) {
				ra = time.Duration(secs) * time.Second
			}
			if ra > d {
				d = ra
			}
		}
	}
	return d
}

// parseError decodes an error body into an APIError. Error payloads are
// small JSON objects carrying a detail/message/error field, but the type is
// populated from the status and raw body even when decoding fails.
func parseError(op string, resp *http.Response, body []byte) *APIError {
	e := &APIError{
		Status:  resp.StatusCode,
		Op:      op,
		Headers: resp.Header.Clone(),
		RawBody: body,
	}
	if len(body) == 0 {
		return e
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(body, &decoded); err == nil {
		for _, key := range []string{"detail", "message", "error", "error_message", "errors"} {
			if raw, ok := decoded[key]; ok {
				var s string
				if json.Unmarshal(raw, &s) == nil {
					e.Message = s
				} else {
					e.Message = string(raw)
				}
				return e
			}
		}
	}
	if len(body) > 300 {
		e.Message = string(body[:300]) + "…"
	} else {
		e.Message = string(body)
	}
	return e
}

// decodeJSON decodes exactly one JSON value into target. Empty bodies and
// trailing values are rejected.
func decodeJSON(body []byte, target interface{}) error {
	if len(body) == 0 {
		return io.ErrUnexpectedEOF
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(target); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return fmt.Errorf("unexpected trailing data: %w", err)
	}
	return nil
}

// decodeResult decodes one response body and unwraps the `result` envelope.
// Most operations wrap rows in a top-level {"result": ...} object; several
// legacy summary/transaction endpoints return a bare object instead — when
// no "result" key exists the whole decoded object is returned.
func decodeResult(op string, body []byte) (interface{}, error) {
	var decoded map[string]interface{}
	if err := decodeJSON(body, &decoded); err != nil {
		return nil, &TransportError{Op: op, Err: err}
	}
	if decoded == nil {
		return nil, &TransportError{Op: op, Err: errors.New("expected a JSON object, got null")}
	}
	if v, ok := decoded["result"]; ok {
		return v, nil
	}
	return decoded, nil
}

// asRows asserts a decoded result payload to a row list. Wide ML/aggregate
// rows stay map[string]any — the column sets are documented per operation in
// the schema descriptions, not as fixed machine-readable schemas.
func asRows(op string, v interface{}) ([]map[string]interface{}, error) {
	list, ok := v.([]interface{})
	if !ok {
		return nil, &TransportError{Op: op, Err: fmt.Errorf("expected a JSON array under \"result\", got %T", v)}
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		row, ok := item.(map[string]interface{})
		if !ok {
			return nil, &TransportError{Op: op, Err: fmt.Errorf("expected object rows under \"result\", got %T", item)}
		}
		out = append(out, row)
	}
	return out, nil
}

// asStrings asserts a decoded catalog payload to a list of symbol strings.
func asStrings(op string, v interface{}) ([]string, error) {
	list, ok := v.([]interface{})
	if !ok {
		return nil, &TransportError{Op: op, Err: fmt.Errorf("expected a JSON string array under \"result\", got %T", v)}
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		value, ok := item.(string)
		if !ok {
			return nil, &TransportError{Op: op, Err: fmt.Errorf("expected string entries under \"result\", got %T", item)}
		}
		out = append(out, value)
	}
	return out, nil
}

// asObject asserts a decoded result payload to a JSON object.
func asObject(op string, v interface{}) (map[string]interface{}, error) {
	obj, ok := v.(map[string]interface{})
	if !ok || obj == nil {
		return nil, &TransportError{Op: op, Err: fmt.Errorf("expected a JSON object, got %T", v)}
	}
	return obj, nil
}

type responseBodyLimitError struct{ limit int64 }

func (e *responseBodyLimitError) Error() string {
	return fmt.Sprintf("response body exceeds %d-byte limit", e.limit)
}

func readResponseBody(r io.Reader, limit int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, limit))
	if err != nil {
		return body, err
	}
	if int64(len(body)) < limit {
		return body, nil
	}
	var extra [1]byte
	n, err := io.ReadFull(r, extra[:])
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if n > 0 {
		return body, &responseBodyLimitError{limit: limit}
	}
	return body, nil
}
