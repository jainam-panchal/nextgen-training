package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRoundRobinPreemptsLowPriorityWhenHighPriorityArrives(t *testing.T) {
	queue := NewRoundRobinQueue()
	rrScheduler := NewRoundRobinScheduler(queue)
	rrScheduler.SetVerbose(false)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue.Push(&Task{
		PID:         1,
		Name:        "low-long-task",
		Priority:    10,
		CPUBurst:    300 * time.Millisecond,
		ArrivalTime: time.Now(),
		Status:      StatusReady,
	})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		rrScheduler.Run(ctx)
	}()

	time.Sleep(60 * time.Millisecond)

	queue.Push(&Task{
		PID:         2,
		Name:        "high-priority-task",
		Priority:    1,
		CPUBurst:    20 * time.Millisecond,
		ArrivalTime: time.Now(),
		Status:      StatusReady,
	})

	deadline := time.After(2 * time.Second)

	for !rrScheduler.Metrics().IsComplete(2) {
		select {
		case <-deadline:
			t.Fatal("round-robin scheduler did not complete tasks before timeout")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	cancel()
	wg.Wait()

	if rrScheduler.Metrics().CompletedTaskCount() != 2 {
		t.Fatalf(
			"expected 2 completed tasks, got %d",
			rrScheduler.Metrics().CompletedTaskCount(),
		)
	}
}
