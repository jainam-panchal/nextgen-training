# CPU Scheduler (Day 10)

This project implements:
- Priority scheduling using a min-heap
- Aging to reduce starvation
- Round-robin scheduling by priority class

## Detailed Docs

For in-depth developer documentation, start here:

- [Docs Index](./docs/INDEX.md)

## Architecture (ASCII)

```text
                          +-----------------------------+
                          |        cmd/sim main         |
                          | flags: tasks, scheduler,    |
                          | profile, quiet, pprof-addr  |
                          +--------------+--------------+
                                         |
                     +-------------------+-------------------+
                     |                                       |
                     v                                       v
          +-----------------------+                +-----------------------+
          |  Priority mode        |                |  Round-Robin mode     |
          +-----------------------+                +-----------------------+
          | TaskProducer          |                | RR task loader        |
          |  -> SafeTaskHeap.Push |                |  -> RoundRobinQueue   |
          +-----------+-----------+                +-----------+-----------+
                      |                                        |
                      v                                        v
        +-----------------------------+            +-----------------------------+
        | SafeTaskHeap                |            | RoundRobinQueue             |
        | - wraps MinHeap[*Task]      |            | - high/medium/low queues    |
        | - Mutex protected           |            | - class-based time quantum   |
        | - notifyCh wakeup signal    |            +-------------+---------------+
        +-------------+---------------+                          |
                      |                                          v
                      v                            +-----------------------------+
        +-----------------------------+            | RoundRobinScheduler         |
        | PriorityScheduler           |            | - Pop by class              |
        | - blocks on notifyCh when   |            | - run for quantum chunk     |
        |   queue empty               |            | - requeue or complete       |
        | - runTask updates metrics   |            +-----------------------------+
        +-------------+---------------+
                      ^
                      |
        +-------------+---------------+
        | AgingService                 |
        | - periodic AgeWaitingTasks() |
        +------------------------------+

Cross-cutting:
- SchedulerMetrics records completion, switches, starvation, wait-time distribution.
- Optional profiling writes cpu/mem/goroutine/mutex profiles.
- pprof HTTP endpoint runs in background for live inspection.
```

## Architecture Notes (Developer View)

- `Task` is the shared scheduling unit. Priority scheduler and RR scheduler both consume `Task`, but queueing semantics differ.
- `SafeTaskHeap` is the core priority queue abstraction: thread-safe heap operations plus wakeup signaling (`notifyCh`) so scheduler does not busy-poll.
- Priority mode concurrency model:
  - producer goroutine pushes tasks
  - scheduler goroutine pops/runs tasks
  - aging goroutine periodically adjusts waiting task priorities
- RR mode concurrency model:
  - queue groups tasks into high/medium/low classes
  - scheduler gives each task a class quantum and may preempt/requeue
- Metrics are centralized in `SchedulerMetrics`, so both schedulers expose consistent operational stats.

## What Was Tested (Simple Note)

- Heap benchmarks test pure heap operations: push, pop, mixed push/pop.
- Scheduler benchmarks test scheduling/queue overhead at scale (1K/10K/50K tasks).
- 10K memory check is run through simulation + heap profile output.
- Profiling artifacts are stored in `artifacts/profiles/`.

## Run

```bash
# default: priority scheduler, 100 tasks, profiling on
go run ./cmd/sim

# fast profiling run (1000 tasks)
go run ./cmd/sim -tasks=1000 -scheduler=priority -profile=true -quiet=true

# 10K memory-focused run
go run ./cmd/sim -tasks=10000 -scheduler=priority -profile=true -quiet=true
```

## Mandatory Profiling Requirement Coverage

### 1. pprof HTTP endpoint

Implemented in `cmd/sim/main.go`:
- `import _ "net/http/pprof"`
- background server: `http.ListenAndServe(":6060", nil)` (configurable with `-pprof-addr`)

### 2. CPU profile during 1000-task simulation

Command used:

```bash
go run ./cmd/sim -tasks=1000 -scheduler=priority -profile=true -quiet=true
go tool pprof -top artifacts/profiles/cpu.prof
```

Key hotspot from current profile:
- `runtime.procyieldAsm` (spin/yield under contention)
- also visible: `TaskProducer.Run`, queue `Mutex.Lock`, heap push/pop path

### 3. Heap profile: memory use at 10K tasks

Command used:

```bash
go run ./cmd/sim -tasks=10000 -scheduler=priority -profile=true -quiet=true
go tool pprof -top artifacts/profiles/mem.prof
```

Observed at current run:
- Total in-use memory in profile: ~`4166.57kB`
- Top contributors:
  - `runtime.mallocgc` ~`2563.80kB`
  - `fmt.Sprintf` + fmt printer allocations from task-name creation/log formatting

### 4. Goroutine profile: leak check

Command used:

```bash
go tool pprof -top artifacts/profiles/goroutine.prof
```

Observed:
- only expected goroutines at shutdown snapshot (main/profile/server wait states)
- no runaway worker goroutines observed from scheduler/aging/producer

### 5. Mutex profile: contention hotspot

Command used:

```bash
go tool pprof -top artifacts/profiles/mutex.prof
```

Observed top contention:
- `sync.(*Mutex).Unlock` / runtime unlock path
- mapped contention source includes:
  - `(*SafeTaskHeap).Pop`
  - `(*PriorityScheduler).Run`
  - producer push path

This matches expected shared-lock pressure around the task heap wrapper.

## Benchmarks

Required benchmark command:

```bash
go test -bench=. -cpuprofile=artifacts/profiles/cpu_bench.prof -memprofile=artifacts/profiles/mem_bench.prof ./scheduler
```

General benchmark command:

```bash
go test -bench=. -benchmem ./...
```

Sample output (recent run):

```text
BenchmarkMinHeapPush/n=10000-16            	    3266	    350265 ns/op	  357626 B/op	      19 allocs/op
BenchmarkMinHeapPop/n=10000-16     	             573	   2256588 ns/op	  357625 B/op	      19 allocs/op
BenchmarkMinHeapInsert10K-16                   	    3643	    312505 ns/op	  357649 B/op	      19 allocs/op

BenchmarkPrioritySchedulingPopOrder/n=50000-16      37	  31656009 ns/op	 7770153 B/op	   50028 allocs/op
BenchmarkAgingServiceOnQueuedTasks/n=10000-16      206	   5831565 ns/op	 1512486 B/op	   10022 allocs/op
BenchmarkRoundRobinQueuePushPop/n=50000-16         100	  14073912 ns/op	 7506112 B/op	   50062 allocs/op
```

## Tests

```bash
go test ./...
go test -race ./...
```

Recent race result:

```text
ok  	cpu-scheduler/heap
ok  	cpu-scheduler/scheduler
```

## Captured Outputs (Attach These)

### 1. Timeline + Metrics (`go run ./cmd/sim -tasks=8 -scheduler=priority -profile=false -quiet=false`)

```text
[PRODUCER] PID=1 Priority=5 Burst=174ms
[T=0.00s] PID=1 (P5) START | [T=0.17s] PID=1 DONE
[PRODUCER] PID=2 Priority=9 Burst=166ms
[T=0.38s] PID=2 (P9) START | [T=0.54s] PID=2 DONE
[PRODUCER] PID=3 Priority=7 Burst=101ms
[T=0.65s] PID=3 (P7) START | [T=0.75s] PID=3 DONE
[PRODUCER] PID=4 Priority=6 Burst=160ms
[T=0.87s] PID=4 (P6) START | [T=1.03s] PID=4 DONE
[PRODUCER] PID=5 Priority=5 Burst=92ms
[T=1.12s] PID=5 (P5) START | [T=1.21s] PID=5 DONE
[PRODUCER] PID=6 Priority=5 Burst=124ms
[T=1.50s] PID=6 (P5) START | [T=1.62s] PID=6 DONE
[T=1.73s] PID=7 (P9) START | [PRODUCER] PID=7 Priority=9 Burst=233ms
[PRODUCER] PID=8 Priority=2 Burst=198ms
[T=1.97s] PID=7 DONE
[T=1.97s] PID=8 (P2) START | [T=2.17s] PID=8 DONE

--- Scheduler Metrics ---
Completed tasks: 8
Context switches: 8
Starvation count: 0
Throughput: 3.69 tasks/sec
Priority 2 average wait: 52.417551ms
Priority 5 average wait: 41.209µs
Priority 6 average wait: 57.018µs
Priority 7 average wait: 41.254µs
Priority 9 average wait: 49.351µs
```

### 2. CPU profile top (`go tool pprof -top artifacts/profiles/cpu.prof`)

```text
File: sim
Build ID: c4d64c58a575033ee7c6383a3e475fe4a269321e
Type: cpu
Time: 2026-05-17 22:58:30 IST
Duration: 33.94ms, Total samples = 30ms (88.39%)
Showing nodes accounting for 30ms, 100% of 30ms total
      flat  flat%   sum%        cum   cum%
      10ms 33.33% 33.33%       10ms 33.33%  internal/runtime/syscall/linux.Syscall6
      10ms 33.33% 66.67%       10ms 33.33%  internal/sync.(*Mutex).lockSlow
      10ms 33.33%   100%       10ms 33.33%  time.runtimeNow
         0     0%   100%       10ms 33.33%  cpu-scheduler/scheduler.(*PriorityScheduler).Run
         0     0%   100%       10ms 33.33%  cpu-scheduler/scheduler.(*PriorityScheduler).runTask
         0     0%   100%       10ms 33.33%  cpu-scheduler/scheduler.(*SafeTaskHeap).Push
         0     0%   100%       10ms 33.33%  cpu-scheduler/scheduler.(*TaskProducer).Run
         0     0%   100%       10ms 33.33%  internal/sync.(*Mutex).Lock (inline)
         0     0%   100%       10ms 33.33%  main.main
         0     0%   100%       10ms 33.33%  main.runPriority.func1
         0     0%   100%       10ms 33.33%  main.runPriority.func2
         0     0%   100%       10ms 33.33%  main.writeProfile
```

### 3. Heap profile top (`go tool pprof -top artifacts/profiles/mem.prof`)

```text
File: sim
Build ID: c4d64c58a575033ee7c6383a3e475fe4a269321e
Type: inuse_space
Time: 2026-05-17 22:58:30 IST
Showing nodes accounting for 4264.42kB, 100% of 4264.42kB total
      flat  flat%   sum%        cum   cum%
 2051.12kB 48.10% 48.10%  2051.12kB 48.10%  runtime.mallocgc
 1184.27kB 27.77% 75.87%  1184.27kB 27.77%  runtime/pprof.StartCPUProfile
  517.02kB 12.12% 87.99%   517.02kB 12.12%  cpu-scheduler/scheduler.(*SchedulerMetrics).RecordWait
  512.01kB 12.01%   100%   512.01kB 12.01%  cpu-scheduler/scheduler.(*TaskProducer).Run
```

### 4. Goroutine profile top (`go tool pprof -top artifacts/profiles/goroutine.prof`)

```text
File: sim
Build ID: c4d64c58a575033ee7c6383a3e475fe4a269321e
Type: goroutine
Time: 2026-05-17 22:58:30 IST
Showing nodes accounting for 2, 100% of 2 total
      flat  flat%   sum%        cum   cum%
         1 50.00% 50.00%          1 50.00%  runtime.goroutineProfileWithLabels
         1 50.00%   100%          1 50.00%  runtime.notetsleepg
```

### 5. Mutex profile top (`go tool pprof -top artifacts/profiles/mutex.prof`)

```text
File: sim
Build ID: c4d64c58a575033ee7c6383a3e475fe4a269321e
Type: delay
Time: 2026-05-17 22:58:30 IST
Showing nodes accounting for 304.17us, 100% of 304.17us total
      flat  flat%   sum%        cum   cum%
  168.80us 55.50% 55.50%   168.80us 55.50%  runtime.unlock (inline)
  135.37us 44.50%   100%   135.37us 44.50%  sync.(*Mutex).Unlock (partial-inline)
```

### 6. Benchmark output (`go test -bench=. -benchmem ./...`)

```text
BenchmarkMinHeapPush/n=100-16      	  718304	      2325 ns/op	    2040 B/op	       8 allocs/op
BenchmarkMinHeapPush/n=1000-16     	   76926	     15342 ns/op	   25208 B/op	      12 allocs/op
BenchmarkMinHeapPush/n=10000-16    	    4293	    278310 ns/op	  357627 B/op	      19 allocs/op
BenchmarkMinHeapPop/n=100-16       	  311305	      7808 ns/op	    2040 B/op	       8 allocs/op
BenchmarkMinHeapPop/n=1000-16      	    6342	    182972 ns/op	   25208 B/op	      12 allocs/op
BenchmarkMinHeapPop/n=10000-16     	     411	   2701879 ns/op	  357625 B/op	      19 allocs/op
BenchmarkMinHeapMixedPushPop/n=1000-16         	   15786	     79533 ns/op	    8184 B/op	      10 allocs/op
BenchmarkMinHeapMixedPushPop/n=10000-16        	     988	   1088331 ns/op	  128248 B/op	      16 allocs/op
BenchmarkMinHeapInsert10K-16                   	    3450	    382270 ns/op	  357649 B/op	      19 allocs/op
BenchmarkPrioritySchedulingPopOrder/n=1000-16         	    3230	    351392 ns/op	  129696 B/op	    1014 allocs/op
BenchmarkPrioritySchedulingPopOrder/n=10000-16        	     229	   5646058 ns/op	 1430568 B/op	   10021 allocs/op
BenchmarkPrioritySchedulingPopOrder/n=50000-16        	      31	  38313131 ns/op	 7770340 B/op	   50028 allocs/op
BenchmarkAgingServiceOnQueuedTasks/n=1000-16          	    2506	    450790 ns/op	  137888 B/op	    1015 allocs/op
BenchmarkAgingServiceOnQueuedTasks/n=10000-16         	     217	   5499850 ns/op	 1512480 B/op	   10022 allocs/op
BenchmarkRoundRobinQueuePushPop/n=1000-16             	    4653	    235653 ns/op	  138178 B/op	    1030 allocs/op
BenchmarkRoundRobinQueuePushPop/n=10000-16            	     480	   2480514 ns/op	 1377056 B/op	   10044 allocs/op
BenchmarkRoundRobinQueuePushPop/n=50000-16            	      73	  14211302 ns/op	 7506260 B/op	   50062 allocs/op
```
