# Benchmark Guide

## What we benchmark

### Heap package

- `BenchmarkMinHeapPush`
- `BenchmarkMinHeapPop`
- `BenchmarkMinHeapMixedPushPop`
- `BenchmarkMinHeapInsert10K` (memory-focused benchmark)

### Scheduler package

- `BenchmarkPrioritySchedulingPopOrder`
- `BenchmarkAgingServiceOnQueuedTasks`
- `BenchmarkRoundRobinQueuePushPop`

## Commands

All benchmarks:

```bash
go test -bench=. -benchmem ./...
```

Per package:

```bash
go test -bench=. -benchmem ./heap
go test -bench=. -benchmem ./scheduler
```

Benchmark with profile output:

```bash
go test -bench=. -cpuprofile=artifacts/profiles/cpu_bench.prof -memprofile=artifacts/profiles/mem_bench.prof ./scheduler
```

## How to interpret quickly

- `ns/op`: latency per operation (lower is better).
- `B/op`: bytes allocated per operation (lower is better).
- `allocs/op`: allocation count per operation (lower is better).

## Current bottleneck patterns

- Heap benchmarks: expected `heapifyUp/heapifyDown` and swap-heavy behavior.
- Scheduler benchmarks:
  - comparator (`LessTask`)
  - lock contention in queue wrappers
  - allocations from benchmark task creation

