package cryptodatadownload

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ioDiscard returns a writer that discards all output.
func ioDiscard() io.Writer { return io.Discard }

// parseRetryAfter parses an HTTP Retry-After header value, which may be
// either a delta-seconds integer or an HTTP-date. It returns the wait
// duration in seconds. Unparseable values return (0, false).
func parseRetryAfter(value string) (seconds int, ok bool) {
	return parseRetryAfterAt(value, time.Now())
}

func parseRetryAfterAt(value string, now time.Time) (seconds int, ok bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if n, err := strconv.Atoi(value); err == nil {
		if n < 0 {
			return 0, false
		}
		return n, true
	}
	if t, err := http.ParseTime(value); err == nil {
		d := t.Sub(now)
		if d < 0 {
			return 0, true
		}
		return int(d.Round(time.Second) / time.Second), true
	}
	return 0, false
}

// retryableStatus reports whether status is one of the retryable codes.
// Every API operation is an idempotent GET, so these are always safe to
// retry; 400/401/404 are permanent by definition.
func retryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}
