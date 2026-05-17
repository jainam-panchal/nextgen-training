package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"cpu-scheduler/scheduler"
)

func main() {
	taskCount := flag.Int("tasks", 100, "number of tasks to simulate")
	schedulerType := flag.String("scheduler", "priority", "scheduler type: priority|rr")
	enableProfile := flag.Bool("profile", true, "enable cpu/memory/goroutine/mutex profile generation")
	quiet := flag.Bool("quiet", false, "disable verbose task execution logs")
	flag.Parse()

	go startPprofServer(":6060")

	stopCPUProfile, err := maybeStartCPUProfile(*enableProfile)
	if err != nil {
		fmt.Printf("could not start CPU profiling: %v\n", err)
		return
	}
	if stopCPUProfile != nil {
		defer stopCPUProfile()
	}

	switch strings.ToLower(*schedulerType) {
	case "priority":
		runPriority(*taskCount, *quiet, *enableProfile)
	case "rr":
		runRoundRobin(*taskCount, *quiet)
	default:
		fmt.Printf("unknown scheduler type %q, expected priority or rr\n", *schedulerType)
		return
	}

	if *enableProfile {
		writeProfile("goroutine.prof", "goroutine")
		writeProfile("mutex.prof", "mutex")
		writeProfile("mem.prof", "heap")
	}
}

func runPriority(taskCount int, quiet bool, profiling bool) {
	queue := scheduler.NewSafeTaskHeap()
	producer := scheduler.NewTaskProducer(queue)
	producer.SetVerbose(!quiet)
	if profiling {
		// Keep profiling runs fast and focused on scheduler/queue behavior.
		producer.SetBurstRange(0, 0)
		producer.SetSleepRange(0, 0)
	}

	priorityScheduler := scheduler.NewPriorityScheduler(queue)
	priorityScheduler.SetVerbose(!quiet)

	agingService := scheduler.NewAgingService(queue, 500*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	var producerWaitGroup sync.WaitGroup
	var schedulerWaitGroup sync.WaitGroup
	var agingWaitGroup sync.WaitGroup

	producerWaitGroup.Add(1)
	go func() {
		defer producerWaitGroup.Done()
		producer.Run(ctx, taskCount)
	}()

	schedulerWaitGroup.Add(1)
	go func() {
		defer schedulerWaitGroup.Done()
		priorityScheduler.Run(ctx)
	}()

	agingWaitGroup.Add(1)
	go func() {
		defer agingWaitGroup.Done()
		agingService.Run(ctx)
	}()

	producerWaitGroup.Wait()
	for !queue.IsEmpty() {
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	schedulerWaitGroup.Wait()
	agingWaitGroup.Wait()

	fmt.Println()
	priorityScheduler.Metrics().Print()
}

func runRoundRobin(taskCount int, quiet bool) {
	queue := scheduler.NewRoundRobinQueue()
	rrScheduler := scheduler.NewRoundRobinScheduler(queue)
	rrScheduler.SetVerbose(!quiet)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for pid := 1; pid <= taskCount; pid++ {
		queue.Push(&scheduler.Task{
			PID:         pid,
			Name:        fmt.Sprintf("rr-task-%d", pid),
			Priority:    (pid % 10) + 1,
			CPUBurst:    50 * time.Millisecond,
			ArrivalTime: time.Now(),
			Status:      scheduler.StatusReady,
		})
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rrScheduler.Run(ctx)
	}()

	timeout := time.After(60 * time.Second)
	for !rrScheduler.Metrics().IsComplete(taskCount) {
		select {
		case <-timeout:
			fmt.Println("timeout: RR simulation did not complete")
			cancel()
			wg.Wait()
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	cancel()
	wg.Wait()

	fmt.Println()
	rrScheduler.Metrics().Print()
}

func maybeStartCPUProfile(enabled bool) (func(), error) {
	if !enabled {
		return nil, nil
	}
	runtime.SetMutexProfileFraction(1)

	cpuProfile, err := os.Create("cpu.prof")
	if err != nil {
		return nil, err
	}

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		_ = cpuProfile.Close()
		return nil, err
	}

	return func() {
		pprof.StopCPUProfile()
		_ = cpuProfile.Close()
	}, nil
}

func startPprofServer(addr string) {
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("pprof server error on %s: %v\n", addr, err)
	}
}

func writeProfile(fileName string, profileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("could not create %s: %v\n", fileName, err)
		return
	}
	defer file.Close()

	if err := pprof.Lookup(profileName).WriteTo(file, 0); err != nil {
		fmt.Printf("could not write %s: %v\n", profileName, err)
	}
}
