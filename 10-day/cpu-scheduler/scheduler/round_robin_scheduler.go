package scheduler

import (
	"context"
	"fmt"
	"time"
)

type RoundRobinScheduler struct {
	queue   *RoundRobinQueue
	metrics *SchedulerMetrics
}

func NewRoundRobinScheduler(queue *RoundRobinQueue) *RoundRobinScheduler {
	return &RoundRobinScheduler{
		queue:   queue,
		metrics: NewSchedulerMetrics(),
	}
}

func (s *RoundRobinScheduler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.metrics.MarkFinished()
			return
		default:
		}

		task, quantum, ok := s.queue.Pop()
		if !ok {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		s.runTask(task, quantum)
	}
}

func (s *RoundRobinScheduler) runTask(task *Task, quantum time.Duration) {
	task.WaitTime += time.Since(task.ArrivalTime)

	task.Status = StatusRunning
	s.metrics.RecordContextSwitch()

	fmt.Printf(
		"[RR][T=%.2fs] PID=%d (P%d) START quantum=%v burst=%v | ",
		time.Since(s.metrics.StartedAt).Seconds(),
		task.PID,
		task.Priority,
		quantum,
		task.CPUBurst,
	)

	runDuration := quantum
	if task.CPUBurst < quantum {
		runDuration = task.CPUBurst
	}

	chunkSize := 10 * time.Millisecond
	remainingRunDuration := runDuration

	for remainingRunDuration > 0 {
		sleepDuration := chunkSize
		if remainingRunDuration < chunkSize {
			sleepDuration = remainingRunDuration
		}

		time.Sleep(sleepDuration)

		task.CPUBurst -= sleepDuration
		remainingRunDuration -= sleepDuration

		if task.CPUBurst <= 0 {
			task.CPUBurst = 0
			task.Status = StatusCompleted

			s.metrics.RecordCompletedTask()
			s.metrics.RecordWait(task.Priority, task.WaitTime)

			fmt.Printf(
				"[RR][T=%.2fs] PID=%d DONE\n",
				time.Since(s.metrics.StartedAt).Seconds(),
				task.PID,
			)

			return
		}

		if task.Priority >= 4 && s.queue.HasHighPriorityTask() {
			task.Status = StatusReady
			task.ArrivalTime = time.Now()
			s.queue.Push(task)

			fmt.Printf(
				"[RR][T=%.2fs] PID=%d PREEMPTED remaining=%v\n",
				time.Since(s.metrics.StartedAt).Seconds(),
				task.PID,
				task.CPUBurst,
			)

			return
		}
	}

	task.Status = StatusReady
	task.ArrivalTime = time.Now()
	s.queue.Push(task)

	fmt.Printf(
		"[RR][T=%.2fs] PID=%d REQUEUED remaining=%v\n",
		time.Since(s.metrics.StartedAt).Seconds(),
		task.PID,
		task.CPUBurst,
	)
}

func (s *RoundRobinScheduler) Metrics() *SchedulerMetrics {
	return s.metrics
}
