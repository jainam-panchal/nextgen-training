# Profiling Guide

## Goals

- Identify hottest CPU paths under scheduler load.
- Measure memory footprint (especially 10K-task scenario).
- Verify no goroutine leaks after shutdown.
- Identify mutex contention hotspots.

## Commands

Run profiling simulation:

```bash
go run ./cmd/sim -tasks=1000 -scheduler=priority -profile=true -quiet=true
```

For 10K memory-focused profile:

```bash
go run ./cmd/sim -tasks=10000 -scheduler=priority -profile=true -quiet=true
```

Inspect profile dumps:

```bash
go tool pprof -top artifacts/profiles/cpu.prof
go tool pprof -top artifacts/profiles/mem.prof
go tool pprof -top artifacts/profiles/goroutine.prof
go tool pprof -top artifacts/profiles/mutex.prof
```

Live pprof endpoint:

```text
http://localhost:6060/debug/pprof/
```

## Reading the results

### CPU

- Expect hot spots in:
  - heap push/pop and comparator path
  - lock/unlock around shared queue
  - producer/task creation under large task counts

### Memory

- 10K run should show:
  - task allocations
  - heap growth allocations
  - some formatting/log allocations if verbose mode is on

### Goroutine

- After clean shutdown:
  - no unbounded worker goroutines
  - expected runtime and pprof server goroutines only

### Mutex

- Expect contention mainly on `SafeTaskHeap` mutex:
  - producer `Push`
  - scheduler `Pop`
  - aging service rebuild operations

## Practical tips

- Keep `-quiet=true` for cleaner profiling signals.
- Run profiles multiple times and compare common hot paths.
- Keep generated `.prof` files in `artifacts/profiles/`, not project root.

