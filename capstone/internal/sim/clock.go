package sim

import (
	"context"
	"time"
)

// Clock drives simulation ticks. Tests can use ManualClock for deterministic control.
type Clock interface {
	C() <-chan Tick
}

// Tick is a monotonic simulation step. N increments by 1 each tick.
type Tick struct {
	N int64
}

// RealClock emits ticks at a fixed wall-clock interval. Stops when ctx is cancelled.
type RealClock struct {
	c    chan Tick
	stop context.CancelFunc
}

// NewRealClock creates a RealClock that emits a Tick every d.
// The goroutine stops when ctx is cancelled.
func NewRealClock(ctx context.Context, d time.Duration) *RealClock {
	ctx, cancel := context.WithCancel(ctx)
	rc := &RealClock{c: make(chan Tick, 1), stop: cancel}
	go func() {
		t := time.NewTicker(d)
		defer t.Stop()
		var n int64
		for {
			select {
			case <-ctx.Done():
				close(rc.c)
				return
			case <-t.C:
				n++
				rc.c <- Tick{N: n}
			}
		}
	}()
	return rc
}

func (r *RealClock) C() <-chan Tick { return r.c }

// ManualClock emits ticks only when Step() is called. Used for deterministic tests.
type ManualClock struct {
	c chan Tick
	n int64
}

// NewManualClock creates a ManualClock with a buffered channel (capacity 1024).
func NewManualClock() *ManualClock {
	return &ManualClock{c: make(chan Tick, 1024)}
}

func (m *ManualClock) C() <-chan Tick { return m.c }

// Step advances the clock by one tick and sends it on the channel.
func (m *ManualClock) Step() {
	m.n++
	m.c <- Tick{N: m.n}
}


