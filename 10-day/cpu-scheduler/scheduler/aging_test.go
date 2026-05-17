package scheduler

import (
	"testing"
	"time"
)

func TestAgingBoostsLowPriorityTask(t *testing.T) {
	queue := NewSafeTaskHeap()

	queue.Push(&Task{
		PID:         1,
		Name:        "low-priority-task",
		Priority:    10,
		CPUBurst:    time.Millisecond,
		ArrivalTime: time.Now(),
		Status:      StatusReady,
	})

	for i := 0; i < 9; i++ {
		queue.AgeWaitingTasks()
	}

	task, err := queue.Peek()
	if err != nil {
		t.Fatalf("expected task, got error: %v", err)
	}

	if task.Priority != 1 {
		t.Fatalf("expected priority 1 after aging, got %d", task.Priority)
	}
}
