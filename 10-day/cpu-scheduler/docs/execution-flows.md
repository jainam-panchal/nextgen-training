# Execution Flows

## Priority Scheduler Flow

1. `cmd/sim` builds `SafeTaskHeap`, `TaskProducer`, `PriorityScheduler`, `AgingService`.
2. Producer generates tasks and pushes them into `SafeTaskHeap`.
3. Each push triggers `notifyCh` (best-effort, buffered).
4. Scheduler loop:
   - tries `Pop()`
   - if empty, blocks on `notifyCh` or context cancellation
   - if task exists, runs task and records metrics
5. Aging loop periodically calls `AgeWaitingTasks()`:
   - drain heap
   - decrease priority value (toward 1) for waiting tasks
   - push tasks back
6. Once producer is done and queue is drained, context is cancelled and goroutines exit.

## Round Robin Flow

1. `cmd/sim` creates `RoundRobinQueue` and `RoundRobinScheduler`.
2. Tasks are inserted into class queues (`high/medium/low`) based on priority.
3. Scheduler loop:
   - pop from highest non-empty class
   - run task in chunks for assigned quantum
   - complete or requeue task with updated burst
   - preempt medium/low task if high-priority work appears
4. When all tasks complete, context is cancelled and scheduler exits.

## Status lifecycle

- Common statuses:
  - `ready`
  - `running`
  - `completed`
  - `starved` (priority scheduler when wait threshold exceeded)

## Cancellation behavior

- Priority scheduler:
  - checks cancellation while waiting on empty queue.
- Round-robin scheduler:
  - checks cancellation each loop tick.

