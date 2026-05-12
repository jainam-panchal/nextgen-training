package main

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

// Benchmark guide:
// - ReadHeavy: pure Get throughput under concurrency.
// - WriteHeavy: pure Set contention behavior.
// - Mixed90Read10Write: common production-like mixed traffic.
// - MixedRatios: crossover study across 99/1 -> 50/50 read/write.
// - ShardSweepMixed90Read10Write: shard-count tuning at fixed mixed workload.
//
// Read metrics:
// - ns/op: latency/throughput proxy (lower is better).
// - B/op and allocs/op: memory overhead per operation.
// - n=...: keyspace size scale for the same workload.

var concurrentSink uint64

const (
	keyPrefix      = "key-"
	defaultKeySize = 25
	defaultShards  = 32

	// Old scale set:
	// concurrencyScale1 = 1_000
	// concurrencyScale2 = 10_000
	// concurrencyScale3 = 100_000
	//
	// New lighter scale set (faster iteration while still showing trends):
	concurrencyScale1 = 500
	concurrencyScale2 = 5_000
	concurrencyScale3 = 50_000
)

var concurrencyScales = []int{concurrencyScale1, concurrencyScale2, concurrencyScale3}

func makeBenchmarkKeys(count int, keySizeBytes int, prefix string) []string {
	const indexWidth = 9

	if keySizeBytes < len(prefix)+indexWidth {
		panic("keySizeBytes is too small")
	}

	keys := make([]string, count)
	for i := range keys {
		key := fmt.Sprintf("%s%09d", prefix, i)
		keys[i] = key + strings.Repeat("x", keySizeBytes-len(key))
	}

	return keys
}

func BenchmarkConcurrentMapReadHeavy(b *testing.B) {
	for _, n := range concurrencyScales {
		keys := makeBenchmarkKeys(n, defaultKeySize, keyPrefix)

		benchmarks := []struct {
			name string
			new  func() ConcurrentMap
		}{
			{name: "LockedBuiltinMap", new: func() ConcurrentMap { return NewLockedBuiltinMap() }},
			{name: fmt.Sprintf("LockedShardedBuiltinMap/shards=%d", defaultShards), new: func() ConcurrentMap { return NewLockedShardedBuiltinMap(defaultShards) }},
			{name: "SyncMap", new: func() ConcurrentMap { return NewSyncMapAdapter() }},
		}

		for _, bm := range benchmarks {
			b.Run(fmt.Sprintf("%s/n=%d", bm.name, n), func(b *testing.B) {
				hashMap := bm.new()
				for keyIndex, key := range keys {
					hashMap.Set(key, keyIndex)
				}

				b.ReportAllocs()
				b.ReportMetric(float64(n), "keys/op")
				b.ResetTimer()

				b.RunParallel(func(pb *testing.PB) {
					keyIndex := 0
					localSum := 0
					for pb.Next() {
						value, _ := hashMap.Get(keys[keyIndex])
						localSum += value
						keyIndex++
						if keyIndex == len(keys) {
							keyIndex = 0
						}
					}
					atomic.AddUint64(&concurrentSink, uint64(localSum))
				})
			})
		}
	}
}

// Measures concurrent write-only behavior to expose lock contention and write overhead.
func BenchmarkConcurrentMapWriteHeavy(b *testing.B) {
	for _, n := range concurrencyScales {
		keys := makeBenchmarkKeys(n, defaultKeySize, keyPrefix)

		benchmarks := []struct {
			name string
			new  func() ConcurrentMap
		}{
			{name: "LockedBuiltinMap", new: func() ConcurrentMap { return NewLockedBuiltinMap() }},
			{name: fmt.Sprintf("LockedShardedBuiltinMap/shards=%d", defaultShards), new: func() ConcurrentMap { return NewLockedShardedBuiltinMap(defaultShards) }},
			{name: "SyncMap", new: func() ConcurrentMap { return NewSyncMapAdapter() }},
		}

		for _, bm := range benchmarks {
			b.Run(fmt.Sprintf("%s/n=%d", bm.name, n), func(b *testing.B) {
				hashMap := bm.new()

				b.ReportAllocs()
				b.ReportMetric(float64(n), "keys/op")
				b.ResetTimer()

				b.RunParallel(func(pb *testing.PB) {
					keyIndex := 0
					for pb.Next() {
						hashMap.Set(keys[keyIndex], keyIndex)
						keyIndex++
						if keyIndex == len(keys) {
							keyIndex = 0
						}
					}
				})
			})
		}
	}
}

func runMixedParallelBenchmark(
	b *testing.B,
	n int,
	newMap func() ConcurrentMap,
	writeEvery int,
) {
	keys := makeBenchmarkKeys(n, defaultKeySize, keyPrefix)
	hashMap := newMap()
	// Preload map so benchmark measures mixed traffic, not cold-start inserts.
	for keyIndex, key := range keys {
		hashMap.Set(key, keyIndex)
	}

	b.ReportAllocs()
	b.ReportMetric(float64(n), "keys/op")
	readPct := 100 - (100 / writeEvery)
	b.ReportMetric(float64(readPct), "read_pct")
	b.ResetTimer()

	var workerCounter atomic.Uint64
	b.RunParallel(func(pb *testing.PB) {
		// Give each worker a different starting point to avoid lock-step access.
		workerID := int(workerCounter.Add(1))
		keyIndex := workerID % len(keys)
		operationIndex := workerID
		localSum := 0

		for pb.Next() {
			key := keys[keyIndex]
			// Every Nth operation is a write; all others are reads.
			if operationIndex%writeEvery == 0 {
				hashMap.Set(key, keyIndex)
			} else {
				value, found := hashMap.Get(key)
				if found {
					localSum += value
				}
			}

			// Stride walk avoids all goroutines hammering exactly same key sequence.
			keyIndex += 17
			if keyIndex >= len(keys) {
				keyIndex %= len(keys)
			}
			operationIndex++
		}

		// Sink prevents compiler from optimizing out read path.
		atomic.AddUint64(&concurrentSink, uint64(localSum))
	})
}

// Measures multiple read/write mixes to find where each map strategy starts to win or lose.
func BenchmarkConcurrentMapMixedRatios(b *testing.B) {
	ratios := []struct {
		name       string
		writeEvery int
	}{
		{name: "99Read1Write", writeEvery: 100},
		{name: "95Read5Write", writeEvery: 20},
		{name: "90Read10Write", writeEvery: 10},
		{name: "80Read20Write", writeEvery: 5},
		{name: "50Read50Write", writeEvery: 2},
	}

	for _, n := range concurrencyScales {
		for _, ratio := range ratios {
			benchmarks := []struct {
				name string
				new  func() ConcurrentMap
			}{
				{name: "LockedBuiltinMap", new: func() ConcurrentMap { return NewLockedBuiltinMap() }},
				{name: fmt.Sprintf("LockedShardedBuiltinMap/shards=%d", defaultShards), new: func() ConcurrentMap { return NewLockedShardedBuiltinMap(defaultShards) }},
				{name: "SyncMap", new: func() ConcurrentMap { return NewSyncMapAdapter() }},
			}

			for _, bm := range benchmarks {
				b.Run(fmt.Sprintf("%s/%s/n=%d", ratio.name, bm.name, n), func(b *testing.B) {
					runMixedParallelBenchmark(b, n, bm.new, ratio.writeEvery)
				})
			}
		}
	}
}

// Sweeps shard counts for the locked sharded map at 90/10 mix to find practical shard tuning.
func BenchmarkLockedShardedMapShardSweepMixed90Read10Write(b *testing.B) {
	// Old shard sweep:
	// shardCounts := []int{1, 2, 4, 8, 16, 32, 64, 128}
	//
	// New lighter sweep:
	shardCounts := []int{1, 4, 16, 32, 64}
	const writeEvery = 10 // 90/10 mix

	for _, n := range concurrencyScales {
		for _, shardCount := range shardCounts {
			b.Run(fmt.Sprintf("n=%d/shards=%d", n, shardCount), func(b *testing.B) {
				runMixedParallelBenchmark(
					b,
					n,
					func() ConcurrentMap { return NewLockedShardedBuiltinMap(shardCount) },
					writeEvery,
				)
			})
		}
	}
}
