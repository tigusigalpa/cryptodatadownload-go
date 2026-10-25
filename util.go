package cryptodatadownload

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

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
	maxSeconds := int(^uint(0) >> 1)
	digits := true
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			digits = false
			break
		}
	}
	if digits {
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil || n > uint64(maxSeconds) {
			return maxSeconds, true
		}
		return int(n), true
	}
	if t, err := http.ParseTime(value); err == nil {
		d := t.Sub(now)
		if d < 0 {
			return 0, true
		}
		seconds := d / time.Second
		if d%time.Second != 0 {
			seconds++
		}
		if int64(seconds) > int64(maxSeconds) {
			return maxSeconds, true
		}
		return int(seconds), true
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
