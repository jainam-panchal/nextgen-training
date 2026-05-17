package scheduler

import (
	"fmt"
	"time"
)

type SchedulerMetrics struct {
	StartedAt       time.Time
	FinishedAt      time.Time
	CompletedTasks  int
	ContextSwitches int
	StarvationCount int

	WaitTimesByPriority map[int][]time.Duration
}

func NewSchedulerMetrics() *SchedulerMetrics {
	now := time.Now()

	return &SchedulerMetrics{
		StartedAt:           now,
		FinishedAt:          now,
		WaitTimesByPriority: make(map[int][]time.Duration),
	}
}

func (m *SchedulerMetrics) RecordWait(priority int, wait time.Duration) {
	m.WaitTimesByPriority[priority] = append(m.WaitTimesByPriority[priority], wait)
}

func (m *SchedulerMetrics) Throughput() float64 {
	elapsed := m.FinishedAt.Sub(m.StartedAt).Seconds()

	if elapsed <= 0 {
		return 0
	}

	return float64(m.CompletedTasks) / elapsed
}

func (m *SchedulerMetrics) AverageWait(priority int) time.Duration {
	waits, exists := m.WaitTimesByPriority[priority]

	if !exists || len(waits) == 0 {
		return 0
	}

	var total time.Duration

	for _, wait := range waits {
		total += wait
	}

	return total / time.Duration(len(waits))
}

func (m *SchedulerMetrics) Print() {
	fmt.Println("\n--- Scheduler Metrics ---")

	fmt.Printf("Completed tasks: %d\n", m.CompletedTasks)
	fmt.Printf("Context switches: %d\n", m.ContextSwitches)
	fmt.Printf("Starvation count: %d\n", m.StarvationCount)
	fmt.Printf("Throughput: %.2f tasks/sec\n", m.Throughput())

	for priority := 1; priority <= 10; priority++ {
		avgWait := m.AverageWait(priority)

		if avgWait > 0 {
			fmt.Printf(
				"Priority %d average wait: %v\n",
				priority,
				avgWait,
			)
		}
	}
}
