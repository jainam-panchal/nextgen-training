package scheduler

import "time"

type TaskStatus int

const (
	StateReady TaskStatus = iota
	StateRunning
	StatusCompleted
	StatusStarved
)

func (s TaskStatus) String() string {
	switch s {
	case StateReady:
		return "ready"
	case StateRunning:
		return "running"
	case StatusCompleted:
		return "completed"
	case StatusStarved:
		return "starved"
	default:
		return "unknown"
	}
}

type Task struct {
	PID         int
	Name        string
	Priority    int // 1 highest, 10 lowest
	CPUBurst    time.Duration
	Deadline    time.Time
	ArrivalTime time.Time
	WaitTime    time.Duration
	Status      TaskStatus
}

func LessTask(first *Task, second *Task) bool {
	if first.Priority != second.Priority {
		return first.Priority < second.Priority
	}

	return first.ArrivalTime.Before(second.ArrivalTime)
}
