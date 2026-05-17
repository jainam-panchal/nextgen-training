package scheduler

import (
	"sync"
	"time"
)

const (
	HighQuantum   = 50 * time.Millisecond
	MediumQuantum = 100 * time.Millisecond
	LowQuantum    = 200 * time.Millisecond
)

type RoundRobinQueue struct {
	mutex sync.Mutex

	highQueue   []*Task
	mediumQueue []*Task
	lowQueue    []*Task
}

func NewRoundRobinQueue() *RoundRobinQueue {
	return &RoundRobinQueue{
		highQueue:   make([]*Task, 0),
		mediumQueue: make([]*Task, 0),
		lowQueue:    make([]*Task, 0),
	}
}

func PriorityClass(priority int) string {
	switch {
	case priority >= 1 && priority <= 3:
		return "high"
	case priority >= 4 && priority <= 6:
		return "medium"
	default:
		return "low"
	}
}

func (q *RoundRobinQueue) Push(task *Task) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	switch PriorityClass(task.Priority) {
	case "high":
		q.highQueue = append(q.highQueue, task)
	case "medium":
		q.mediumQueue = append(q.mediumQueue, task)
	default:
		q.lowQueue = append(q.lowQueue, task)
	}
}

func (q *RoundRobinQueue) Pop() (*Task, time.Duration, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.highQueue) > 0 {
		task := q.highQueue[0]
		q.highQueue = q.highQueue[1:]
		return task, HighQuantum, true
	}

	if len(q.mediumQueue) > 0 {
		task := q.mediumQueue[0]
		q.mediumQueue = q.mediumQueue[1:]
		return task, MediumQuantum, true
	}

	if len(q.lowQueue) > 0 {
		task := q.lowQueue[0]
		q.lowQueue = q.lowQueue[1:]
		return task, LowQuantum, true
	}

	return nil, 0, false
}

func (q *RoundRobinQueue) HasHighPriorityTask() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.highQueue) > 0
}

func (q *RoundRobinQueue) IsEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.highQueue) == 0 &&
		len(q.mediumQueue) == 0 &&
		len(q.lowQueue) == 0
}
