package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cpu-scheduler/heap"
)

type PriorityScheduler struct {
	queue   *SafeTaskHeap
	metrics *SchedulerMetrics
}

func NewPriorityScheduler(queue *SafeTaskHeap) *PriorityScheduler {
	return &PriorityScheduler{
		queue:   queue,
		metrics: NewSchedulerMetrics(),
	}
}

func (s *PriorityScheduler) Run(ctx context.Context) {
	for {
		task, err := s.queue.Pop()
		if err != nil {
			if !errors.Is(err, heap.ErrEmptyHeap) {
				fmt.Printf("scheduler pop error: %v\n", err)
			}

			select {
			case <-ctx.Done():
				s.metrics.MarkFinished()
				return
			default:
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}

		if task == nil {
			fmt.Println("scheduler received nil task")
			time.Sleep(10 * time.Millisecond)
			continue
		}

		s.runTask(task)
	}
}

func (s *PriorityScheduler) runTask(task *Task) {
	now := time.Now()
	task.WaitTime = now.Sub(task.ArrivalTime)

	if task.WaitTime > 5*time.Second {
		task.Status = StatusStarved
		s.metrics.RecordCompletedTask()
	}

	task.Status = StatusRunning
	s.metrics.RecordContextSwitch()

	fmt.Printf("[T=%.2fs] PID=%d (P%d) START | ",
		time.Since(s.metrics.StartedAt).Seconds(),
		task.PID,
		task.Priority,
	)
	time.Sleep(task.CPUBurst) // working.. duuh duh duh

	task.Status = StatusCompleted
	s.metrics.RecordCompletedTask()
	s.metrics.RecordWait(task.Priority, task.WaitTime)

	fmt.Printf("[T=%.2fs] PID=%d DONE\n",
		time.Since(s.metrics.StartedAt).Seconds(),
		task.PID,
	)
}

func (s *PriorityScheduler) Metrics() *SchedulerMetrics {
	return s.metrics
}
