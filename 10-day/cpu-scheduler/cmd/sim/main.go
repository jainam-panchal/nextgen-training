package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"cpu-scheduler/scheduler"
)

func main() {
	cpuProfile, err := os.Create("cpu.prof")
	if err != nil {
		fmt.Printf("could not create CPU profile: %v\n", err)
		return
	}
	defer cpuProfile.Close()

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		fmt.Printf("could not start CPU profile: %v\n", err)
		return
	}
	defer pprof.StopCPUProfile()

	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	queue := scheduler.NewSafeTaskHeap()
	producer := scheduler.NewTaskProducer(queue)
	priorityScheduler := scheduler.NewPriorityScheduler(queue)
	agingService := scheduler.NewAgingService(queue, 500*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	var producerWaitGroup sync.WaitGroup
	var schedulerWaitGroup sync.WaitGroup
	var agingWaitGroup sync.WaitGroup

	producerWaitGroup.Add(1)
	go func() {
		defer producerWaitGroup.Done()
		producer.Run(ctx, 100)
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

	writeProfile("mem.prof", "heap")
	writeProfile("goroutine.prof", "goroutine")
	writeProfile("mutex.prof", "mutex")

	fmt.Println()
	priorityScheduler.Metrics().Print()
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
