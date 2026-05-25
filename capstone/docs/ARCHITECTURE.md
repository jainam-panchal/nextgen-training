# Architecture Overview

## Project Structure

```
capstone/
├── cmd/cli/main.go         — Scenario-driven CLI runner
├── cmd/server/main.go       — REST API + pprof server
├── internal/
│   ├── engine/              — Core simulation engine
│   │   ├── types.go         — Types, constants, config
│   │   └── engine.go        — Everything: signals, vehicles, congestion, API
│   ├── dijkstra/            — Shortest path routing
│   │   └── dijkstra.go
│   ├── model/               — Graph data structures
│   │   ├── topology.go       — Graph, Road, Edge types + demo generator
│   │   └── topology_safe.go — Thread-safe wrapper (not used by engine currently)
│   ├── scenario/            — Scenario JSON loader + builder
│   │   └── scenario.go
│   ├── httpapi/             — REST API handlers
│   │   └── router.go
│   ├── ds/                  — Data structures
│   │   ├── priority_queue.go — Generic priority queue (min-heap)
│   │   └── ring_buffer.go    — Sliding window ring buffer
│   └── sim/                 — Simulation clock
│       └── clock.go
├── scenarios/               — JSON scenario files
│   ├── 01_basic.json
│   ├── 02_emergency_preempt.json
│   └── 03_grid.json
└── README.md
```

## Packages — What Each One Does

| Package | Role |
|---------|------|
| `cmd/cli` | **CLI entrypoint.** Loads a scenario JSON, creates engine, registers vehicle, drives ticks, prints output. |
| `cmd/server` | **Server entrypoint.** Creates engine with demo graph, mounts REST API + pprof, graceful shutdown. |
| `internal/engine` | **The brain.** All simulation logic lives here: tick driver, signals, vehicles, congestion, emergency, public API. |
| `internal/dijkstra` | **Router.** `ShortestPath()` using a generic priority queue. Weight function: `minutes = distance/speed * 60 * (1 + (congestion-1)/10)`. |
| `internal/model` | **Data types.** `Graph` (adjacency list), `Road`, `Edge`. `GraphReader` interface for thread-safe reading. |
| `internal/scenario` | **Config loader.** Parses scenario JSON (verbose or compact road format, grid shorthand), validates, builds `model.Graph`. |
| `internal/httpapi` | **API handlers.** Thin JSON wrappers over `engine.Engine` methods. |
| `internal/ds` | **Utility structs.** Generic `PriorityQueue[T]` (min-heap via `container/heap`), `RingBuffer[T]` (sliding window). |
| `internal/sim` | **Clock abstraction.** `RealClock` (ticks via `time.Ticker`), `ManualClock` (test ticks via `.Step()`). Shared `Tick{N}` type. |

## Entry Points — What Happens at Startup

### CLI (`cmd/cli/main.go`)

```
1. Parse -scenario and -max-ticks flags
2. scenario.LoadFile(path) → parses JSON, validates, returns *Scenario
3. s.Build() → sorts intersection IDs, builds model.Graph from scenario roads
4. engine.New(cfg, graph, clock) → creates engine, indexes roads, inits signals
5. engine.Start(ctx) → launches goroutines:
     a) tick driver goroutine (clock → cond broadcast)
     b) one goroutine per intersection (signal cycling)
     c) congestion updater goroutine (periodic)
6. engine.RegisterVehicle() → computes route, creates vehicle, launches goroutine
7. Main loop:
     a) WaitNextTick() → blocks until next simulation tick
     b) Process scheduled events (emergency dispatch)
     c) WaitAllProcessed() → waits for ALL goroutines to finish tick N
     d) VehicleStatus() → read current state
     e) Print state-change line
```

### Server (`cmd/server/main.go`)

```
1. Parse -addr and -tick flags
2. model.GenerateDemoGraph() → 20-node graph with chain + shortcuts
3. engine.New() + engine.Start() (same as CLI)
4. httpapi.New(e) → wires routes
5. Mount pprof handlers at /debug/pprof/
6. http.ListenAndServe with graceful shutdown
```

## Data Flow Diagram

```
Scenario JSON ──> scenario.LoadFile() ──> Scenario struct
                                               │
                                          s.Build()
                                               │
                                               v
                                          model.Graph
                                               │
                                    engine.New(cfg, graph, clock)
                                               │
                                          engine.Start()
                                               │
                         ┌─────────────────────┼─────────────────────┐
                         │                     │                     │
                         v                     v                     v
              Tick Driver Goroutine    N × Intersection       Congestion Updater
              (broadcasts tickN via    Goroutines             Goroutine
               sync.Cond)              (cycle signal phases)  (update road congestion)
                         │
                         │  ticks broadcast via cond
                         │
          ┌──────────────┼──────────────┐
          v              v              v
    Vehicle        Vehicle        Vehicle
    Goroutine 1    Goroutine 2    Goroutine N
    (step: move   (step: wait    (step: arrive)
     or wait)      or move)
          │
          └── model.GraphReader (via graphView)
              └── reads current congestion from engine's road state
              └── avoids emergency corridor roads (penalty)
```

## Key Design Decisions

### 1. Tick-based Simulation (not wall-clock)

The entire simulation runs on **integer ticks**. A tick is a discrete time step — like a turn in a board game. The real wall-clock duration of a tick is `Config.TickDuration` (e.g. 50ms for fast CLI, 1s for server), but internally all logic uses tick numbers.

**Why?** Determinism. Ticks make it easy to write tests (`ManualClock.Step()` exactly controls time), avoid race conditions (each goroutine processes one tick at a time), and reason about timing.

### 2. sync.Cond for Tick Broadcast (not channels)

All goroutines (signals, vehicles, congestion updater) need to run on every tick. A single channel cannot safely broadcast to N consumers — only one goroutine would receive each tick. Using `sync.Cond`, the tick driver broadcasts to ALL waiting goroutines at once.

```
Tick Driver:                        Waiters (signals, vehicles, congestion):
  tickMu.Lock()                       tickMu.Lock()
  tickN = nextN                       for tickN < last+1:
  tickCond.Broadcast()                  tickCond.Wait()
  tickMu.Unlock()                     tickMu.Unlock()
                                      process tick
```

### 3. Progress Barrier (WaitAllProcessed)

After the CLI advances the tick, it needs to know when ALL goroutines have finished processing that tick before reading state. The `WaitAllProcessed()` method tracks the highest tick each goroutine has completed via `markIntersectionProcessed()`, `markVehicleProcessed()`, and `markCongestionProcessed()`.

### 4. graphView — Dynamic Weight Reader

The engine implements `model.GraphReader` via `graphView`. When Dijkstra calls `Neighbors(node)`, the `graphView` reads the **current** congestion levels from the engine's `roadByID` map (under `roadMu.RLock()`). This means every shortest-path query gets live congestion data, including:
- Normal congestion (`1..10`)
- Emergency corridor penalty (congestion=10, distance×50)

### 5. Road Weights are Minutes

`dijkstra.WeightMinutes(road)` computes travel time in **minutes**:
```
base = distance_km / speed_kmph * 60       ← time at speed limit
mult = 1 + (congestion - 1) / 10           ← congestion multiplier
weight = base * mult
```

The result is in minutes, and the conversion to ticks is `ceil(weight)` with a minimum of 1 tick. This means each hop takes at least 1 tick.

### 6. Column-Major Grid IDs

Grid scenarios use **column-major** ID ordering: IDs increase vertically down each column, then wrap to the next column. This makes `movementDirection(from, to)` simple:
- `abs(to - from) == 1` → NS (vertical)
- otherwise → EW (horizontal)

```
Column 0    Column 1    Column 2
    0           4           8
    1           5           9
    2           6          10
    3           7          11

Path 0→4: diff=4 → EW (horizontal)
Path 0→1: diff=1 → NS (vertical)
```

## Goroutine Lifecycle

```
engine.Start()
  │
  ├── Tick Driver Goroutine
  │     └── reads from clock.C(), broadcasts tickN, runs until ctx cancelled
  │
  ├── Intersection Goroutine (× NumIntersections)
  │     └── waits for each tick, calls intersection.step() (advance phase or remain)
  │
  └── Congestion Updater Goroutine
        └── waits for each tick, updates all road congestion, clears expired corridor

engine.RegisterVehicle()
  └── Vehicle Goroutine
        └── waits for each tick, calls vehicle.step() (move/wait/arrive)

engine.Stop() / ctx cancelled
  └── tick driver closes → cond.Broadcast wakes all → goroutines exit
  └── sync.WaitGroup waits for all
```
