# Data Structures

## `heap.MinHeap[T]`

Generic binary min-heap with custom comparator:
- `Push`: append + heapify-up
- `Pop`: root swap/remove + heapify-down
- `Peek`, `Len`, `IsEmpty`, `Verify`, `Items`

Complexity:
- `Push`: `O(log n)`
- `Pop`: `O(log n)`
- `Peek`: `O(1)`

## `scheduler.SafeTaskHeap`

Thread-safe wrapper around `MinHeap[*Task]`:
- `mutex` guards all heap operations
- `notifyCh` signals scheduler on new tasks

Why it exists:
- heap itself is not goroutine-safe
- scheduler needs blocking wake-up without polling

## `scheduler.RoundRobinQueue`

Three class queues:
- high (priority 1-3)
- medium (priority 4-6)
- low (priority 7-10)

Each class has fixed quantum:
- high: 50ms
- medium: 100ms
- low: 200ms

Current implementation detail:
- uses slice front-pop (`q = q[1:]`) which is simple but can create copy pressure at scale.

## `scheduler.Task`

Core scheduling unit:
- `PID`, `Name`, `Priority`
- `CPUBurst`, `ArrivalTime`, `WaitTime`
- `Status`
- `Deadline` currently present for future policies, not hard-enforced by current schedulers.

## `scheduler.SchedulerMetrics`

Thread-safe metrics aggregator:
- completed tasks
- context switches
- starvation count
- per-priority wait distributions
- throughput from start/finish timestamps

