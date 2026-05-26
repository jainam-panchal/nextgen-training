package engine

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

// BenchmarkMutexVsRWMutex compares sync.Mutex and sync.RWMutex under
// a read-heavy workload that mimics the engine's road-map access pattern:
// multiple concurrent goroutines reading road weights, with occasional writes.
//
// Run: go test -bench=BenchmarkMutexVsRWMutex -benchmem ./internal/engine/

func BenchmarkMutexVsRWMutex(b *testing.B) {
	nRoads := 100
	data := make(map[int]int, nRoads)
	for i := 0; i < nRoads; i++ {
		data[i] = rand.Intn(10) + 1
	}

	readers := []int{1, 5, 10, 50}
	writeRatio := []float64{0.01, 0.05, 0.1}

	for _, nr := range readers {
		for _, wr := range writeRatio {
			b.Run(fmt.Sprintf("readers=%d_writes=%.0f%%", nr, wr*100), func(b *testing.B) {
				// --- Mutex ---
				b.Run("Mutex", func(b *testing.B) {
					var mu sync.Mutex
					var reads, writes atomic.Int64
					ops := int64(0)

					b.SetParallelism(nr)
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							if rand.Float64() < wr {
								mu.Lock()
								data[rand.Intn(nRoads)] = rand.Intn(10) + 1
								mu.Unlock()
								writes.Add(1)
							} else {
								mu.Lock()
								_ = data[rand.Intn(nRoads)]
								mu.Unlock()
								reads.Add(1)
							}
							ops++
						}
					})
					_ = ops
				})

				// --- RWMutex ---
				b.Run("RWMutex", func(b *testing.B) {
					var mu sync.RWMutex
					var reads, writes atomic.Int64
					ops := int64(0)

					b.SetParallelism(nr)
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							if rand.Float64() < wr {
								mu.Lock()
								data[rand.Intn(nRoads)] = rand.Intn(10) + 1
								mu.Unlock()
								writes.Add(1)
							} else {
								mu.RLock()
								_ = data[rand.Intn(nRoads)]
								mu.RUnlock()
								reads.Add(1)
							}
							ops++
						}
					})
					_ = ops
				})
			})
		}
	}
}
