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
	verbose bool
}

func NewPriorityScheduler(queue *SafeTaskHeap) *PriorityScheduler {
	return &PriorityScheduler{
		queue:   queue,
		metrics: NewSchedulerMetrics(),
		verbose: true,
	}
}

func (s *PriorityScheduler) SetVerbose(verbose bool) {
	s.verbose = verbose
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
			case <-s.queue.NotifyChan():
				continue
			}
		}

		if task == nil {
			if s.verbose {
				fmt.Println("scheduler received nil task")
			}
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
		s.metrics.RecordStarvation()
	}

	task.Status = StatusRunning
	s.metrics.RecordContextSwitch()

	if s.verbose {
		fmt.Printf("[T=%.2fs] PID=%d (P%d) START | ",
			time.Since(s.metrics.StartedAt).Seconds(),
			task.PID,
			task.Priority,
		)
	}
	time.Sleep(task.CPUBurst) // working.. duuh duh duh

	task.Status = StatusCompleted
	s.metrics.RecordCompletedTask()
	s.metrics.RecordWait(task.Priority, task.WaitTime)

	if s.verbose {
		fmt.Printf("[T=%.2fs] PID=%d DONE\n",
			time.Since(s.metrics.StartedAt).Seconds(),
			task.PID,
		)
	}
}

func (s *PriorityScheduler) Metrics() *SchedulerMetrics {
	return s.metrics
}
