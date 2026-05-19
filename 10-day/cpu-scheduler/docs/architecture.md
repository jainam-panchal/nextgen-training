# Architecture

## High-level view

```text
                +--------------------------------------+
                |              cmd/sim                 |
                | flags: tasks, scheduler, profile,   |
                | quiet                                |
                +------------------+-------------------+
                                   |
              +--------------------+--------------------+
              |                                         |
              v                                         v
   +-------------------------+              +---------------------------+
   | Priority Scheduler Path |              | Round Robin Scheduler Path|
   +-------------------------+              +---------------------------+
   | TaskProducer            |              | Task Loader (in main)     |
   |   -> SafeTaskHeap.Push  |              |   -> RoundRobinQueue.Push |
   +------------+------------+              +------------+--------------+
                |                                        |
                v                                        v
   +-----------------------------+            +-----------------------------+
   | SafeTaskHeap                |            | RoundRobinQueue             |
   | - Mutex-protected MinHeap   |            | - high/medium/low queues    |
   | - notifyCh wake-up signal   |            | - class-based quanta        |
   +-------------+---------------+            +-------------+---------------+
                 |                                          |
                 v                                          v
   +-----------------------------+            +-----------------------------+
   | PriorityScheduler           |            | RoundRobinScheduler         |
   | - Pop best task             |            | - Pop by class              |
   | - Execute task              |            | - Execute for quantum       |
   | - Record metrics            |            | - Preempt/requeue if needed |
   +-------------+---------------+            +-------------+---------------+
                 ^
                 |
   +-------------+---------------+
   | AgingService                 |
   | - periodic priority boost    |
   +------------------------------+
```

## Concurrency model

- Priority mode runs 3 goroutines:
  - producer
  - scheduler
  - aging service
- Round-robin mode runs:
  - scheduler (tasks are preloaded in current implementation)
- Shutdown is coordinated through `context.Context`.

## Synchronization strategy

- `SafeTaskHeap` uses `sync.Mutex` around heap operations.
- Wake-up is event-driven via `notifyCh`:
  - producer push emits signal
  - scheduler blocks on signal if queue is empty
- `RoundRobinQueue` uses `sync.Mutex` around queue slices.

## Metrics and observability

- `SchedulerMetrics` tracks:
  - completed task count
  - context switch count
  - starvation count
  - wait-time distribution by priority
  - throughput from start/end timestamps
- Profiling support:
  - pprof HTTP endpoint (`net/http/pprof`)
  - CPU, heap, goroutine, mutex profile dumps

