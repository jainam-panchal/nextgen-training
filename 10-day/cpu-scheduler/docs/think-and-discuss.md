# Think & Discuss: Scheduler Design Deep Dive

This document answers the design-review questions for the Day 10 scheduler project in a practical, systems-oriented way.

---

## 1. How this compares to Linux CFS

### Your current scheduler

- Priority scheduler uses:
  - discrete priority levels (`1..10`)
  - min-heap ordered by `(priority, arrivalTime)`
  - optional aging to reduce starvation
- Behavior is class/priority driven, not fairness-driven.

### Linux CFS (Completely Fair Scheduler)

- CFS tries to approximate "ideal fair sharing" of CPU over time.
- It does **not** primarily schedule by fixed priority buckets.
- Core concept is **virtual runtime** (`vruntime`):
  - task that has run less (smaller vruntime) gets CPU first.
  - tasks that run accumulate vruntime and move "rightward."
- CFS runqueue is a **red-black tree** keyed by `vruntime`.
  - Left-most node is next task to run.
  - Insert/remove/rebalance are `O(log n)`.

### Data structure comparison

- Your priority min-heap:
  - Great for "highest-priority-first."
  - Not directly fair over long horizons.
  - No stable notion of "consumed fair share."
- CFS red-black tree:
  - Good for ordered selection by continuously changing key (`vruntime`).
  - Makes fairness and preemption decisions smoother.

### If you want a CFS-like model in this project

Use a task field like:
- `VirtualRuntime float64`

Schedule the smallest `VirtualRuntime`, then after each run:
- `task.VirtualRuntime += execTime * weightFactor(priority)`

At that point, using a balanced BST (or ordered structure) is conceptually closer than fixed-priority heap scheduling.

---

## 2. Aging goroutine writes while scheduler reads: synchronization needs

### Current concurrency risk

- Scheduler pops tasks concurrently.
- Aging goroutine mutates priorities in queued tasks.
- Producer pushes concurrently.

This is shared mutable state across goroutines.

### Why synchronization is required

Without locking:
- data races on heap internal slice
- invalid heap ordering during concurrent mutation
- potential memory corruption/panic

### Why `sync.RWMutex` alone is not "the answer"

`RWMutex` is a lock primitive, not a correctness proof.

For this specific structure:

- Heap operations that mutate shape (`Push`, `Pop`, reheapify) are writes.
- Aging operation (`AgeWaitingTasks`) is also a write-heavy operation:
  - drains heap
  - mutates each task priority
  - pushes back
- So most hot operations are write paths, not read paths.

Result:
- `RWMutex` gives little benefit because write lock is still dominant.
- A read lock cannot be used for scheduler `Pop` or aging because both mutate.

So the important part is:
- strict mutual exclusion around heap mutation (which your `sync.Mutex` already enforces),
- plus careful wake-up coordination (`notifyCh`) to avoid busy polling.

---

## 3. Real-time tasks with hard deadlines

### Requirement shape

If task **must** finish before deadline, priority-only scheduling is not enough.

### Recommended algorithm: EDF (Earliest Deadline First)

- Dynamic-priority algorithm.
- Always pick task with smallest absolute deadline.
- Optimal for single-CPU preemptive scheduling (under classical assumptions).

### Data structure

Use min-heap keyed by `deadline`:
- `less(a,b) => a.Deadline.Before(b.Deadline)`

### Required features in implementation

1. Admission control (optional but important for hard RT):
- reject task if system cannot feasibly schedule it.

2. Preemption:
- if new task arrives with earlier deadline than running task, preempt.

3. Deadline miss handling:
- classify as missed, trigger policy:
  - drop
  - continue best-effort
  - escalate alert

4. Time accounting:
- need precise remaining execution time (`remaining burst`) and wakeups.

### Practical note

Your current architecture can evolve into EDF by:
- adding an `RT` queue with deadline-keyed heap,
- integrating preemption decision in run loop,
- keeping metrics for deadline misses.

---

## 4. Context switch overhead (simulation vs real OS)

### In your simulation

"Context switch" currently is a logical metric increment (`RecordContextSwitch()`).

What it includes:
- bookkeeping + scheduler loop transitions

What it does **not** include:
- CPU register save/restore
- TLB/cache disruption
- kernel/user mode transitions
- runqueue balancing across cores

So your context-switch overhead is intentionally lightweight and not hardware-realistic.

### In real OS

A context switch has real microarchitectural costs:

- save/restore CPU context
- scheduler decision + runqueue locks
- cache locality loss
- branch predictor/TLB effects
- cross-core migration costs

Latency varies by workload and platform, but can materially impact throughput in high-switch-rate workloads.

### How to model it better in simulation (if needed)

- Add configurable switch penalty (`switchCost`) to scheduler loop.
- Track:
  - switches per second
  - cumulative switch time
  - % overhead vs task execution time

That gives a better performance discussion even if still not kernel-accurate.

---

## 5. Task dependencies (B after A): DAG model

### Correct abstraction

Dependencies form a **DAG (Directed Acyclic Graph)**:

- Node = task
- Edge `A -> B` means B depends on A

### Data structures

1. Adjacency list:
- `dependents[A] = [B, C, ...]`

2. In-degree map:
- `inDegree[task] = number of unmet prerequisites`

3. Ready queue:
- tasks with `inDegree == 0`

### Runtime flow

1. Build DAG.
2. Enqueue all zero in-degree tasks.
3. Scheduler runs ready tasks.
4. On task completion:
   - for each dependent `d`:
     - decrement `inDegree[d]`
     - if becomes zero, enqueue `d`
5. Detect cycle if tasks remain with non-zero in-degree and ready queue empty.

### Integration with your schedulers

Use dependency resolution layer before queueing:
- dependency manager decides when task is "ready"
- then pushes to `SafeTaskHeap` / `RoundRobinQueue`

This keeps scheduler logic clean and makes dependencies orthogonal to policy.

---

## Suggested next evolution path (if you continue this project)

1. Add dependency manager (DAG + in-degree) feeding existing queues.
2. Add optional EDF mode using task deadlines.
3. Add switch-cost simulation parameter for more realistic scheduling tradeoffs.
4. Compare fairness metrics between current priority+aging and a CFS-like vruntime prototype.

