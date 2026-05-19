package scheduler

import (
	"fmt"
	"sync"
	"time"
)

type SchedulerMetrics struct {
	mutex sync.Mutex

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

func (m *SchedulerMetrics) RecordContextSwitch() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContextSwitches++
}

func (m *SchedulerMetrics) RecordCompletedTask() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CompletedTasks++
}

func (m *SchedulerMetrics) RecordStarvation() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.StarvationCount++
}

func (m *SchedulerMetrics) RecordWait(priority int, wait time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.WaitTimesByPriority[priority] =
		append(m.WaitTimesByPriority[priority], wait)
}

func (m *SchedulerMetrics) MarkFinished() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.FinishedAt = time.Now()
}

func (m *SchedulerMetrics) IsComplete(totalTasks int) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.CompletedTasks >= totalTasks
}

func (m *SchedulerMetrics) CompletedTaskCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.CompletedTasks
}

func (m *SchedulerMetrics) Throughput() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	elapsed := m.FinishedAt.Sub(m.StartedAt).Seconds()
	if elapsed <= 0 {
		return 0
	}

	return float64(m.CompletedTasks) / elapsed
}

func (m *SchedulerMetrics) AverageWait(priority int) time.Duration {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.averageWaitLocked(priority)
}

func (m *SchedulerMetrics) averageWaitLocked(priority int) time.Duration {
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
	m.mutex.Lock()
	defer m.mutex.Unlock()

	elapsed := m.FinishedAt.Sub(m.StartedAt).Seconds()
	throughput := 0.0
	if elapsed > 0 {
		throughput = float64(m.CompletedTasks) / elapsed
	}

	fmt.Println("\n--- Scheduler Metrics ---")
	fmt.Printf("Completed tasks: %d\n", m.CompletedTasks)
	fmt.Printf("Context switches: %d\n", m.ContextSwitches)
	fmt.Printf("Starvation count: %d\n", m.StarvationCount)
	fmt.Printf("Throughput: %.2f tasks/sec\n", throughput)

	for priority := 1; priority <= 10; priority++ {
		waits, exists := m.WaitTimesByPriority[priority]
		if !exists || len(waits) == 0 {
			continue
		}

		fmt.Printf(
			"Priority %d average wait: %v\n",
			priority,
			m.averageWaitLocked(priority),
		)
	}
}
