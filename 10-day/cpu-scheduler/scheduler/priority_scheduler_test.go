package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPrioritySchedulerHighPriorityRunsFirst(t *testing.T) {
	queue := NewSafeTaskHeap()
	now := time.Now()

	for pid := 1; pid <= 10; pid++ {
		queue.Push(&Task{
			PID:         pid,
			Name:        "low-priority-task",
			Priority:    10,
			CPUBurst:    time.Millisecond,
			ArrivalTime: now,
			Status:      StatusReady,
		})
	}

	queue.Push(&Task{
		PID:         99,
		Name:        "high-priority-task",
		Priority:    1,
		CPUBurst:    time.Millisecond,
		ArrivalTime: now.Add(time.Second),
		Status:      StatusReady,
	})

	task, err := queue.Pop()
	if err != nil {
		t.Fatalf("expected task, got error: %v", err)
	}

	if task.PID != 99 {
		t.Fatalf("expected high-priority PID 99 first, got PID %d", task.PID)
	}
}

func TestPrioritySchedulerWakesWhenTaskArrives(t *testing.T) {
	queue := NewSafeTaskHeap()
	scheduler := NewPriorityScheduler(queue)
	scheduler.SetVerbose(false)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scheduler.Run(ctx)
	}()

	time.Sleep(30 * time.Millisecond)

	queue.Push(&Task{
		PID:         1,
		Name:        "wake-test",
		Priority:    5,
		CPUBurst:    time.Millisecond,
		ArrivalTime: time.Now(),
		Status:      StatusReady,
	})

	deadline := time.After(2 * time.Second)
	for !scheduler.Metrics().IsComplete(1) {
		select {
		case <-deadline:
			t.Fatal("scheduler did not wake and process pushed task")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	cancel()
	wg.Wait()
}

func TestPrioritySchedulerStopsWhileWaitingOnEmptyQueue(t *testing.T) {
	queue := NewSafeTaskHeap()
	scheduler := NewPriorityScheduler(queue)
	scheduler.SetVerbose(false)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		scheduler.Run(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}
