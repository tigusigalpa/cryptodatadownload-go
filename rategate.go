package cryptodatadownload

import (
	"context"
	"sync"
	"time"
)

// rateGate is a minimum-interval gate applied before every request attempt.
// It enforces a steady-state requests-per-second ceiling (default 1 rps) so
// callers stay under the API's burst (60/min) and sustained (10K/24h) limits.
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

// wait blocks until the next request may be sent or ctx is cancelled. A rate
// of <= 0 disables gating entirely.
func (g *rateGate) wait(ctx context.Context) error {
	if g == nil || g.rps <= 0 {
		return nil
	}
	for {
		g.mu.Lock()
		now := g.clock.Now()
		if g.next.Before(now) {
			g.next = now
		}
		fire := g.next
		g.next = fire.Add(time.Duration(float64(time.Second) / g.rps))
		g.mu.Unlock()

		d := fire.Sub(now)
		if d <= 0 {
			return nil
		}
		if err := g.sleeper.Sleep(ctx, d); err != nil {
			return err
		}
		// After a context-aware sleep, re-evaluate: another goroutine may have
		// taken our slot or the clock may not have advanced as far as we slept.
	}
}
