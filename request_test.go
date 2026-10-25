package cryptodatadownload

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func transportTestClient(t *testing.T, transport http.RoundTripper, opts ...Option) *Client {
	t.Helper()
	options := append([]Option{
		WithHTTPClient(&http.Client{Transport: transport}),
		WithRequestsPerSecond(0),
		WithRetryMaxAttempts(1),
	}, opts...)
	c, err := NewClient("test-key", options...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

type brokenBody struct{}

func (brokenBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (brokenBody) Close() error             { return nil }

func TestResponseFailuresRetryPolicy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   func() io.ReadCloser
		want   int
	}{
		{"oversized success", 200, func() io.ReadCloser { return io.NopCloser(strings.NewReader(strings.Repeat("x", 65))) }, 1},
		{"server error independent of success limit", 503, func() io.ReadCloser { return io.NopCloser(strings.NewReader(strings.Repeat("x", 65))) }, 3},
		{"broken validation error", 400, func() io.ReadCloser { return brokenBody{} }, 1},
		{"broken auth error", 401, func() io.ReadCloser { return brokenBody{} }, 1},
		{"broken not found", 404, func() io.ReadCloser { return brokenBody{} }, 1},
		{"broken success", 200, func() io.ReadCloser { return brokenBody{} }, 3},
		{"broken server error", 503, func() io.ReadCloser { return brokenBody{} }, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0
			c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
				attempts++
				return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: tc.body()}, nil
			}), WithRetryMaxAttempts(3), WithMaxResponseBodyBytes(64), WithSleeper(sleeperFunc(func(context.Context, time.Duration) error { return nil })))
			_, err := c.BinanceOHLC.Spot(context.Background(), SpotParams{})
			if err == nil || (tc.status == 200 && AsTransport(err) == nil) || (tc.status != 200 && AsAPIError(err) == nil) {
				t.Fatalf("error = %v, want error matching the response status", err)
			}
			if attempts != tc.want {
				t.Errorf("attempts = %d, want %d", attempts, tc.want)
			}
		})
	}
}

func TestBrokenResponseHonorsRetryAfter(t *testing.T) {
	var waits []time.Duration
	c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {"10"}}, Body: brokenBody{}}, nil
	}), WithRetryMaxAttempts(2), WithSleeper(sleeperFunc(func(_ context.Context, d time.Duration) error {
		waits = append(waits, d)
		return nil
	})))
	_, _ = c.BinanceOHLC.Spot(context.Background(), SpotParams{})
	if len(waits) != 1 || waits[0] != 10*time.Second {
		t.Fatalf("waits = %v, want [10s]", waits)
	}
}

type interruptedReader struct{ calls int }

func (r *interruptedReader) Read(p []byte) (int, error) {
	r.calls++
	switch r.calls {
	case 1:
		return copy(p, "abc"), nil
	case 2:
		return 0, nil
	case 3:
		return copy(p, "d"), nil
	default:
		return 0, io.EOF
	}
}

func TestReadResponseBodyDetectsOversizeAfterEmptyRead(t *testing.T) {
	if _, err := readResponseBody(&interruptedReader{}, 3); err == nil {
		t.Fatal("oversized body accepted after a zero-length read")
	}
}

func TestRetryAfterNeverWaitsTooLittle(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 800_000_000, time.UTC)
	value := now.Add(200 * time.Millisecond).Format(http.TimeFormat)
	if seconds, ok := parseRetryAfterAt(value, now); !ok || seconds != 1 {
		t.Fatalf("Retry-After = (%d, %v), want (1, true)", seconds, ok)
	}
	c, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	d := c.backoffWithRetryAfter(1, http.Header{"Retry-After": {"9223372037"}})
	expected := time.Duration(math.MaxInt64)
	if strconv.IntSize == 32 {
		expected = time.Duration(math.MaxInt32) * time.Second
	}
	if d != expected {
		t.Errorf("large Retry-After = %s, want saturated duration", d)
	}
}

func TestNullObjectResponseIsRejected(t *testing.T) {
	for _, body := range []string{"null", `{"result":null}`} {
		c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
		}))
		_, err := c.Catalog.CftcCotAvailable(context.Background(), CftcCotAvailableParams{})
		if err == nil {
			t.Errorf("body %s returned success", body)
		}
	}
}

func TestCancellationAfterRetryWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	attempts := 0
	c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, io.ErrUnexpectedEOF
	}), WithRetryMaxAttempts(3), WithSleeper(sleeperFunc(func(context.Context, time.Duration) error {
		cancel()
		return nil
	})))
	_, err := c.BinanceOHLC.Spot(ctx, SpotParams{})
	if !errors.Is(err, context.Canceled) || attempts != 1 {
		t.Fatalf("error = %v, attempts = %d; want context.Canceled after one attempt", err, attempts)
	}
}
