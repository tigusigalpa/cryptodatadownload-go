package cryptodatadownload

import (
	"context"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRequestsPerSecondValidation(t *testing.T) {
	for _, rps := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1), math.SmallestNonzeroFloat64, math.MaxFloat64} {
		if _, err := NewClient("test-key", WithRequestsPerSecond(rps)); err == nil {
			t.Errorf("rps=%v accepted", rps)
		}
	}
	for _, rps := range []float64{0, 0.5, 1, 1000} {
		if _, err := NewClient("test-key", WithRequestsPerSecond(rps)); err != nil {
			t.Errorf("rps=%v rejected: %v", rps, err)
		}
	}
}

type limiterFunc func(context.Context) error

func (f limiterFunc) Wait(ctx context.Context) error { return f(ctx) }

func TestSharedLimiterCoordinatesConcurrentClients(t *testing.T) {
	var waits atomic.Int32
	limiter := limiterFunc(func(ctx context.Context) error {
		waits.Add(1)
		return ctx.Err()
	})
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"result":[]}`))}, nil
	})
	first := transportTestClient(t, transport, WithRateLimiter(limiter))
	second := transportTestClient(t, transport, WithRateLimiter(limiter))
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := first
			if i%2 != 0 {
				c = second
			}
			if _, err := c.BinanceOHLC.Spot(context.Background(), SpotParams{}); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if waits.Load() != 20 {
		t.Fatalf("shared limiter calls = %d, want 20", waits.Load())
	}
}
