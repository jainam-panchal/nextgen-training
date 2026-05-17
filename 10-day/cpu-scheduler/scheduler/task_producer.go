package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type TaskProducer struct {
	queue       *SafeTaskHeap
	verbose     bool
	minBurstMs  int
	maxBurstMs  int
	minSleepMs  int
	maxSleepMs  int
	priorityMax int
}

func NewTaskProducer(queue *SafeTaskHeap) *TaskProducer {
	return &TaskProducer{
		queue:       queue,
		verbose:     true,
		minBurstMs:  50,
		maxBurstMs:  250,
		minSleepMs:  100,
		maxSleepMs:  400,
		priorityMax: 10,
	}
}

func (p *TaskProducer) SetVerbose(verbose bool) {
	p.verbose = verbose
}

func (p *TaskProducer) SetBurstRange(minMs int, maxMs int) {
	if minMs < 0 {
		minMs = 0
	}
	if maxMs < minMs {
		maxMs = minMs
	}
	p.minBurstMs = minMs
	p.maxBurstMs = maxMs
}

func (p *TaskProducer) SetSleepRange(minMs int, maxMs int) {
	if minMs < 0 {
		minMs = 0
	}
	if maxMs < minMs {
		maxMs = minMs
	}
	p.minSleepMs = minMs
	p.maxSleepMs = maxMs
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
			Priority:    rand.Intn(p.priorityMax) + 1,
			CPUBurst:    time.Duration(randomInRange(p.minBurstMs, p.maxBurstMs)) * time.Millisecond,
			Deadline:    time.Now().Add(10 * time.Second),
			ArrivalTime: time.Now(),
			Status:      StatusReady,
		}

		p.queue.Push(task)

		if p.verbose {
			fmt.Printf(
				"[PRODUCER] PID=%d Priority=%d Burst=%v\n",
				task.PID,
				task.Priority,
				task.CPUBurst,
			)
		}

		sleepDuration := time.Duration(randomInRange(p.minSleepMs, p.maxSleepMs)) * time.Millisecond

		time.Sleep(sleepDuration)
	}
}

func randomInRange(minInclusive int, maxInclusive int) int {
	if maxInclusive <= minInclusive {
		return minInclusive
	}
	return rand.Intn(maxInclusive-minInclusive+1) + minInclusive
}
