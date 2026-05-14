package leaktest

import (
	"runtime"
	"testing"
	"time"
)

func startLeakyWorker() {
	ch := make(chan int)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- 42 // blocks forever: nobody receives
	}()
}

func TestDetectGoroutineLeak(t *testing.T) {
	runtime.GC()
	baseline := runtime.NumGoroutine()

	startLeakyWorker()

	// Wait long enough for the goroutine to reach the blocked send.
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	after := runtime.NumGoroutine()

	if after > baseline {
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)

		t.Fatalf(
			"goroutine leak detected: before=%d after=%d\n\n%s",
			baseline,
			after,
			buf[:n],
		)
	}
}
