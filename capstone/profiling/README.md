# Profiling — PS Section 6

All flame graphs and plots are SVG files (open in any browser, pan/zoom enabled).

## 1. Dijkstra Scaling Benchmark

| Nodes | ns/op |
|-------|-------|
| 20    | 6,256 |
| 100   | 60,155 |
| 500   | 574,095 |

Scaling plot: [`dijkstra_scaling.svg`](dijkstra_scaling.svg)

O(N log N) — 20→500 is ~25× nodes, ~92× time. Dominated by heap operations (PushItem, PopItem).

## 2. CPU Profile

[`dijkstra_cpu_flame.svg`](dijkstra_cpu_flame.svg) — Dijkstra CPU:
- `ShortestPath`: 47%
- GC (`mallocgc`): 27%
- Heap ops (`heap.Pop`/`heap.Push`): 16%
- `WeightMinutes` (inline): 5%

## 3. Memory Profile

[`dijkstra_mem_flame.svg`](dijkstra_mem_flame.svg) — Dijkstra allocs:
- `ShortestPath`: 99% of all allocations (866MB in all-pairs benchmark)
- `PushItem`: 51% (24B priority queue items)
- `NewPriorityQueue`: 3.2%

## 4. Goroutine Profile

[`goroutine_flame.svg`](goroutine_flame.svg) — 29 goroutines at steady state:
- 20 intersection goroutines (`waitTick` + signal cycle)
- 1 congestion updater
- 1 tick driver
- 1 HTTP server
- 1 main

All goroutines shut down cleanly (verified via `TestLoad100VehiclesRaceSafe` + graceful shutdown).

## 5. Block Profile

[`block_flame.svg`](block_flame.svg) — Blocking time:
- `sync.Mutex.Lock`: 66% — contention during vehicle registration (Dijkstra + road lock)
- `sync.Cond.Wait`: 32% — goroutines parked waiting for tick broadcast
- `sync.RWMutex.Lock`: 4.4%

## 6. Mutex vs RWMutex

[`mutex_1pct.svg`](mutex_1pct.svg) — 1% writes: RWMutex ~20% faster
[`mutex_5pct.svg`](mutex_5pct.svg) — 5% writes: RWMutex ~10-30% faster
[`mutex_10pct.svg`](mutex_10pct.svg) — 10% writes: RWMutex ~15-50% faster

RWMutex wins on read-heavy workloads. At 1% writes with 1 reader the difference is marginal (~212 vs ~323 ns/op). At 50 readers the gap narrows as OS scheduler contention dominates.

## 7. Full Engine Profiles

[`memory_flame.svg`](memory_flame.svg) — Server heap: per-vehicle and per-intersection memory footprint.
