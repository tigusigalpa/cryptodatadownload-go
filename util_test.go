package cryptodatadownload

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestRetryAfterSaturatesUnrepresentableSeconds(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, value := range []string{"9223372036854775808", "99999999999999999999999999999999999"} {
		if seconds, ok := parseRetryAfterAt(value, now); !ok || seconds != maxInt {
			t.Errorf("Retry-After %q = (%d, %v), want saturated seconds", value, seconds, ok)
		}
	}
	value := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC).Format(http.TimeFormat)
	seconds, ok := parseRetryAfterAt(value, now)
	if !ok || seconds <= 0 || (strconv.IntSize == 32 && seconds != maxInt) {
		t.Errorf("future Retry-After overflowed: (%d, %v)", seconds, ok)
	}
}
