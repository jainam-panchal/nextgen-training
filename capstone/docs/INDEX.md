# Capstone Docs Index

Start here to understand the Smart City Traffic Management PoC.

## Quick Start

| Doc | What it covers |
|-----|----------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Project structure, packages, data flow diagram, goroutine lifecycle, key design decisions |
| [ENGINE_FLOW.md](ENGINE_FLOW.md) | Tick-by-tick simulation flow, line-by-line walkthrough of every core function |
| [CONCEPTS.md](CONCEPTS.md) | 13 key concepts explained: ticks, sync.Cond, Dijkstra, signals, preemption, congestion, thread safety |

## How to Read

1. **Start with ARCHITECTURE.md** — understand the project layout, what each package does, and how they connect.
2. **Read ENGINE_FLOW.md** — follow the simulation loop from Start() through each tick cycle.
3. **Use CONCEPTS.md as reference** — look up specific topics (ticks, signals, emergency, etc.) as needed.

## File Map

```
capstone/
├── cmd/cli/main.go           ← start here for CLI flow
├── cmd/server/main.go         ← start here for server flow
├── internal/engine/
│   ├── types.go              ← config, constants (Phase, VehicleState, VehicleType)
│   └── engine.go             ← everything else (1000+ lines)
├── internal/dijkstra/
│   └── dijkstra.go           ← ShortestPath + WeightMinutes
├── internal/model/
│   └── topology.go           ← Graph, Road, Edge, GraphReader interface
├── internal/scenario/
│   └── scenario.go           ← JSON loading, validation, graph building
├── internal/httpapi/
│   └── router.go             ← REST handlers + RunHTTPServer
├── internal/ds/
│   ├── priority_queue.go     ← generic heap
│   └── ring_buffer.go        ← sliding window
├── internal/sim/
│   └── clock.go              ← RealClock, ManualClock, Tick
├── scenarios/
│   ├── 01_basic.json         ← 4-node chain, normal vehicle
│   ├── 02_emergency_preempt.json  ← 2×2 grid, emergency at tick 5
│   └── 03_grid.json          ← 5×4 grid, normal vehicle
└── docs/
    ├── INDEX.md              ← this file
    ├── ARCHITECTURE.md
    ├── ENGINE_FLOW.md
    └── CONCEPTS.md
```

## Key Files by Topic

**Want to understand the simulation engine?**
→ `internal/engine/engine.go`

**Want to see how routing works?**
→ `internal/dijkstra/dijkstra.go` + `internal/ds/priority_queue.go`

**Want to add a new scenario?**
→ `internal/scenario/scenario.go` + look at existing files in `scenarios/`

**Want to understand the API?**
→ `internal/httpapi/router.go`

**Want to understand the grid ID scheme?**
→ `internal/scenario/scenario.go:118-152` (generateGrid function)
