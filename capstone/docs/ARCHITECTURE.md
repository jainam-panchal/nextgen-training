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
│   │   └── topology.go       — Graph, Road, Edge types, GraphReader, demo generator
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
     c) VehicleStatus() → read current vehicle state
     d) Detect state change (suppress repeats)
     e) Print one-line state-change output
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

### 3. Per-Vehicle Processed Tick (WaitVehicleProcessed)

Tests need to synchronize with a specific vehicle's goroutine. Rather than a full barrier (which we removed), each vehicle sets `v.processedTick` atomically after each `step()`. Tests call `WaitVehicleProcessed(plate, tickN)` which busy-waits until the vehicle has processed that tick. This is lighter than a condvar-based barrier and sufficient for test determinism.

### 4. graphView — Dynamic Weight Reader

The engine implements `model.GraphReader` via `graphView`. When Dijkstra calls `Neighbors(node)`, the `graphView` reads the **current** congestion levels from the engine's `roadByID` map (under `roadMu.RLock()`). This means every shortest-path query gets live congestion data, including:
- Normal congestion (`1..10`)
- Emergency corridor penalty (congestion=10, distance×50)

### 5. No Periodic Reroute (Simplified Step)

The vehicle state machine no longer has periodic reroute (`ReRouteEveryTicks` was removed), no `nextIntersection`/`from`/`lastRouteTick` fields, and no `procCond` barrier. The route path is fixed at registration — only the per-hop travel time adapts to live congestion via `roadForHop()`. This simplifies the step function from ~100 lines to ~60 and removes 3 fields from the vehicle struct.

### 6. Road Weights are Minutes

`dijkstra.WeightMinutes(road)` computes travel time in **minutes**:
```
base = distance_km / speed_kmph * 60       ← time at speed limit
mult = 1 + (congestion - 1) / 10           ← congestion multiplier
weight = base * mult
```

The result is in minutes, and the conversion to ticks is `ceil(weight)` with a minimum of 1 tick. This means each hop takes at least 1 tick.

### 7. Column-Major Grid IDs

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
engine.Start() → launches:
  │
  ├── Tick Driver Goroutine (1)
  │     └── reads from clock.C(), broadcasts tickN via sync.Cond
  │
  ├── Intersection Goroutines (× NumIntersections)
  │     └── waitTick → intersection.step() (signal phase advance)
  │
  └── Congestion Updater Goroutine (1)
        └── waitTick → updateCongestion() (recalc all roads, clear expired corridor)

engine.RegisterVehicle() → launches:
  └── Vehicle Goroutine (1 per vehicle)
        └── waitTick → vehicle.step() (move/wait/arrive) → store processedTick

engine.Stop() / ctx cancelled:
  └── tick driver exits → cond.Broadcast wakes all → goroutines see tickClosed → return
  └── sync.WaitGroup.Wait() — all goroutines confirmed done
```
