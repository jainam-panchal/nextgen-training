package scheduler

import (
	"context"
	"time"
)

type AgingService struct {
	queue    *SafeTaskHeap
	interval time.Duration
}

func NewAgingService(queue *SafeTaskHeap, interval time.Duration) *AgingService {
	return &AgingService{
		queue:    queue,
		interval: interval,
	}
}

func (a *AgingService) Run(ctx context.Context) {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.queue.AgeWaitingTasks()
		}
	}
}
