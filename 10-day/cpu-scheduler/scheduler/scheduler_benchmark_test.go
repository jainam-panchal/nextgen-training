package scheduler

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func benchmarkTask(pid int) *Task {
	return &Task{
		PID:         pid,
		Name:        "bench-task",
		Priority:    rand.Intn(10) + 1,
		CPUBurst:    0,
		ArrivalTime: time.Unix(0, 0),
		Status:      StatusReady,
	}
}

func BenchmarkPrioritySchedulingPopOrder(b *testing.B) {
	for _, size := range []int{1000, 10000, 50000} {
		b.Run("n="+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				queue := NewSafeTaskHeap()
				for pid := 1; pid <= size; pid++ {
					queue.Push(benchmarkTask(pid))
				}
				for !queue.IsEmpty() {
					_, _ = queue.Pop()
				}
			}
		})
	}
}

func BenchmarkAgingServiceOnQueuedTasks(b *testing.B) {
	for _, size := range []int{1000, 10000} {
		b.Run("n="+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				queue := NewSafeTaskHeap()
				for pid := 1; pid <= size; pid++ {
					queue.Push(benchmarkTask(pid))
				}
				queue.AgeWaitingTasks()
			}
		})
	}
}

func BenchmarkRoundRobinQueuePushPop(b *testing.B) {
	for _, size := range []int{1000, 10000, 50000} {
		b.Run("n="+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				queue := NewRoundRobinQueue()
				for pid := 1; pid <= size; pid++ {
					queue.Push(benchmarkTask(pid))
				}
				for !queue.IsEmpty() {
					_, _, _ = queue.Pop()
				}
			}
		})
	}
}
