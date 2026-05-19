# Code Map

## Entry points

- `cmd/sim/main.go`
  - main simulator with flags, profiling, and pprof endpoint.
- `cmd/rr-sim/main.go`
  - focused RR demo simulation.

## Scheduler core

- `scheduler/task.go`
  - task model, priority comparator, status enum.
- `scheduler/metrics.go`
  - thread-safe metric recording and reporting.

### Priority path

- `scheduler/priority_scheduler.go`
  - priority scheduler execution loop and task runtime behavior.
- `scheduler/safe_task_heap.go`
  - mutex-protected heap wrapper + wake-up channel.
- `scheduler/aging.go`
  - periodic priority aging.
- `scheduler/task_producer.go`
  - task generation and enqueue pipeline.

### Round-robin path

- `scheduler/round_robin_queue.go`
  - class queues and quanta mapping.
- `scheduler/round_robin_scheduler.go`
  - quantum-based execution, preemption, and requeue logic.

## Heap package

- `heap/min_heap.go`
  - generic min-heap implementation.
- `heap/min_heap_test.go`
  - correctness tests.
- `heap/min_heap_benchmark_test.go`
  - heap performance benchmarks.

## Test and benchmark coverage

- Scheduler tests:
  - `priority_scheduler_test.go`
  - `aging_test.go`
  - `round_robin_scheduler_test.go`
- Scheduler benchmarks:
  - `scheduler_benchmark_test.go`

