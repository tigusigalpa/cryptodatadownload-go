package cryptodatadownload

import (
	"context"
	"sync"
	"time"

	"github.com/tigusigalpa/cryptodatadownload-go/internal/clock"
)

// rateGate is a minimum-interval gate applied before every request attempt.
// It enforces a steady-state requests-per-second ceiling (default 1 rps) so
// callers can pace requests against the API's burst (60/min) limit. Daily
// and endpoint-specific budgets must be enforced by the application.
// It uses the client's injectable clock and sleeper, so tests run without
// real delays. It is safe for concurrent use.
type rateGate struct {
	mu      sync.Mutex
	rps     float64
	clock   Clock
	sleeper Sleeper
	next    time.Time // earliest instant the next request may fire
}

func newRateGate(rps float64, cl Clock, s Sleeper) *rateGate {
	return &rateGate{rps: rps, clock: cl, sleeper: s}
}

// NewRateLimiter creates a concurrency-safe minimum-interval limiter that can
// be shared by clients using one token. Zero disables pacing. It coordinates
// this process only and does not track daily or endpoint-specific budgets.
func NewRateLimiter(requestsPerSecond float64) (RateLimiter, error) {
	if err := validateRequestsPerSecond(requestsPerSecond); err != nil {
		return nil, err
	}
	return newRateGate(requestsPerSecond, clock.SystemClock{}, clock.SystemSleeper{}), nil
}

// Wait implements RateLimiter.
func (g *rateGate) Wait(ctx context.Context) error { return g.wait(ctx) }

// wait blocks until the next request may be sent or ctx is cancelled. A rate
// of <= 0 disables gating entirely.
func (g *rateGate) wait(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if g == nil || g.rps <= 0 {
		return nil
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		g.mu.Lock()
		now := g.clock.Now()
		d := g.next.Sub(now)
		if d <= 0 {
			g.next = now.Add(time.Duration(float64(time.Second) / g.rps))
			g.mu.Unlock()
			return nil
		}
		g.mu.Unlock()
		if err := g.sleeper.Sleep(ctx, d); err != nil {
			return err
		}
		// Only reserve a slot when it can be used. Cancelled waiters do not
		// consume capacity; concurrent callers recheck the next send time.
	}
}
