package cryptodatadownload

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tigusigalpa/cryptodatadownload-go/internal/clock"
)

type sleeperFunc func(context.Context, time.Duration) error

func (f sleeperFunc) Sleep(ctx context.Context, d time.Duration) error { return f(ctx, d) }

func TestRateGateSequentialCalls(t *testing.T) {
	cl := clock.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	var sleeps []time.Duration
	gate := newRateGate(1, cl, sleeperFunc(func(_ context.Context, d time.Duration) error {
		sleeps = append(sleeps, d)
		if len(sleeps) > 4 {
			return errors.New("rate gate keeps reserving slots after sleeping")
		}
		cl.Advance(d)
		return nil
	}))
	for i := 0; i < 5; i++ {
		if err := gate.wait(context.Background()); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
	}
	if len(sleeps) != 4 {
		t.Fatalf("sleeps = %v, want four waits", sleeps)
	}
	for _, d := range sleeps {
		if d != time.Second {
			t.Errorf("sleep = %s, want 1s", d)
		}
	}
}

func TestRateGateCancellationDoesNotReserveSlot(t *testing.T) {
	cl := clock.NewFakeClock(time.Now())
	gate := newRateGate(1, cl, sleeperFunc(func(ctx context.Context, _ time.Duration) error {
		return context.Canceled
	}))
	if err := gate.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := gate.wait(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled wait = %v", err)
	}
	var waited time.Duration
	gate.sleeper = sleeperFunc(func(_ context.Context, d time.Duration) error {
		waited += d
		if waited > time.Second {
			return errors.New("cancelled call consumed a future slot")
		}
		cl.Advance(d)
		return nil
	})
	if err := gate.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRateGateRejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, rps := range []float64{0, 1} {
		gate := newRateGate(rps, clock.SystemClock{}, clock.SystemSleeper{})
		if err := gate.wait(ctx); !errors.Is(err, context.Canceled) {
			t.Errorf("rps=%v: wait = %v, want context.Canceled", rps, err)
		}
	}
}

func TestRateGateConcurrentCalls(t *testing.T) {
	const calls = 12
	gate := newRateGate(1000, clock.SystemClock{}, clock.SystemSleeper{})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	errs := make(chan error, calls)
	var wg sync.WaitGroup
	for i := 0; i < calls; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gate.wait(ctx)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if elapsed := time.Since(start); elapsed < (calls-1)*time.Millisecond {
		t.Errorf("%d calls completed in %s, faster than the configured ceiling", calls, elapsed)
	}
}
