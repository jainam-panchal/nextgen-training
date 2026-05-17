package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type TaskProducer struct {
	queue *SafeTaskHeap
}

func NewTaskProducer(queue *SafeTaskHeap) *TaskProducer {
	return &TaskProducer{
		queue: queue,
	}
}

func (p *TaskProducer) Run(
	ctx context.Context,
	totalTasks int,
) {
	for pid := 1; pid <= totalTasks; pid++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		task := &Task{
			PID:         pid,
			Name:        fmt.Sprintf("task-%d", pid),
			Priority:    rand.Intn(10) + 1,
			CPUBurst:    time.Duration(rand.Intn(200)+50) * time.Millisecond,
			Deadline:    time.Now().Add(10 * time.Second),
			ArrivalTime: time.Now(),
			Status:      StatusReady,
		}

		p.queue.Push(task)

		fmt.Printf(
			"[PRODUCER] PID=%d Priority=%d Burst=%v\n",
			task.PID,
			task.Priority,
			task.CPUBurst,
		)

		sleepDuration :=
			time.Duration(rand.Intn(300)+100) * time.Millisecond

		time.Sleep(sleepDuration)
	}
}
