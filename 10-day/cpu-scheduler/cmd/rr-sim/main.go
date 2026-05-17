package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cpu-scheduler/scheduler"
)

func main() {
	const totalTasks = 2

	queue := scheduler.NewRoundRobinQueue()
	rrScheduler := scheduler.NewRoundRobinScheduler(queue)

	ctx, cancel := context.WithCancel(context.Background())

	queue.Push(&scheduler.Task{
		PID:         1,
		Name:        "low-long-task",
		Priority:    10,
		CPUBurst:    500 * time.Millisecond,
		ArrivalTime: time.Now(),
		Status:      scheduler.StatusReady,
	})

	var schedulerWaitGroup sync.WaitGroup

	schedulerWaitGroup.Add(1)
	go func() {
		defer schedulerWaitGroup.Done()
		rrScheduler.Run(ctx)
	}()

	time.Sleep(60 * time.Millisecond)

	queue.Push(&scheduler.Task{
		PID:         2,
		Name:        "high-priority-task",
		Priority:    1,
		CPUBurst:    50 * time.Millisecond,
		ArrivalTime: time.Now(),
		Status:      scheduler.StatusReady,
	})

	timeout := time.After(5 * time.Second)

	for !rrScheduler.Metrics().IsComplete(totalTasks) {
		select {
		case <-timeout:
			fmt.Println("timeout: RR simulation did not complete")
			cancel()
			schedulerWaitGroup.Wait()
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	cancel()
	schedulerWaitGroup.Wait()

	fmt.Println()
	rrScheduler.Metrics().Print()
}
