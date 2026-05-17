package scheduler

import (
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
