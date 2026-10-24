package cryptodatadownload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// apiResponse is a decoded REST response. The body is fully buffered; the
// underlying connection has already been closed when do returns.
type apiResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

// do performs a GET request against the CryptoDataDownload API, applying
// Token authentication, the client-side rate limiter, retry policy and error
// mapping. path is the operation path relative to the base URL (e.g.
// "/data/ohlc/binance/spot/"). The API key is sent as
// "Authorization: Token <key>" and never appears in the URL, errors or logs.
func (c *Client) do(ctx context.Context, op, path string, query url.Values) (*apiResponse, error) {
	if ctx == nil {
		return nil, &TransportError{Op: op, Err: errors.New("nil context")}
	}
	if err := ctx.Err(); err != nil {
		return nil, &TransportError{Op: op, Err: err}
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
		if err := c.gate.wait(ctx); err != nil {
			return nil, &TransportError{Op: op, Err: err}
		}

		resp, sendErr := c.sendOnce(ctx, fullURL)
		if sendErr != nil {
			lastErr = &TransportError{Op: op, Err: sendErr}
			if attempt < attempts {
				d := c.backoff(attempt)
				c.logRetry(attempt, attempts, d)
				if waitErr := c.wait(ctx, d); waitErr != nil {
					return nil, &TransportError{Op: op, Err: waitErr}
				}
				continue
			}
			return nil, lastErr
		}

		body, readErr := readResponseBody(resp.Body, c.cfg.maxResponseBodyBytes)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = &TransportError{Op: op, StatusCode: resp.StatusCode, Err: readErr}
			if attempt < attempts {
				d := c.backoff(attempt)
				c.logRetry(attempt, attempts, d)
				if waitErr := c.wait(ctx, d); waitErr != nil {
					return nil, &TransportError{Op: op, Err: waitErr}
				}
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return &apiResponse{statusCode: resp.StatusCode, header: resp.Header, body: body}, nil
		}

		apiErr := parseError(op, resp, body)
		lastErr = apiErr

		if !retryableStatus(resp.StatusCode) || attempt >= attempts {
			return nil, apiErr
		}
		d := c.backoffWithRetryAfter(attempt, resp.Header)
		c.logRetry(attempt, attempts, d)
		if waitErr := c.wait(ctx, d); waitErr != nil {
			return nil, &TransportError{Op: op, Err: waitErr}
		}
	}
	return nil, lastErr
}

// sendOnce performs a single HTTP GET without retry logic.
func (c *Client) sendOnce(ctx context.Context, fullURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+c.apiKey)
	req.Header.Set("User-Agent", c.cfg.userAgent)

	return c.cfg.httpClient.Do(req)
}

// buildURL joins the base URL with the operation path and encoded query.
func (c *Client) buildURL(path string, query url.Values) string {
	u := strings.TrimSuffix(c.cfg.baseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
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
			ra := time.Duration(secs) * time.Second
			if ra > d {
				d = ra
			}
		}
	}
	return d
}

// parseError decodes an error body into an ApiError. Error payloads are
// small JSON objects carrying a detail/message/error field, but the type is
// populated from the status and raw body even when decoding fails.
func parseError(op string, resp *http.Response, body []byte) *ApiError {
	e := &ApiError{
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

// asObject asserts a decoded result payload to a JSON object.
func asObject(op string, v interface{}) (map[string]interface{}, error) {
	obj, ok := v.(map[string]interface{})
	if !ok {
		return nil, &TransportError{Op: op, Err: fmt.Errorf("expected a JSON object, got %T", v)}
	}
	return obj, nil
}

func readResponseBody(r io.Reader, limit int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, limit))
	if err != nil {
		return nil, err
	}
	var extra [1]byte
	n, err := r.Read(extra[:])
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if n > 0 {
		return nil, fmt.Errorf("response body exceeds %d-byte limit", limit)
	}
	return body, nil
}
