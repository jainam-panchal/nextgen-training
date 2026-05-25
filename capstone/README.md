## Smart City Traffic Management System (PoC)

This folder contains a working proof-of-concept implementation of the capstone problem statement.

It is intentionally learning-focused:
- clear data-structure mapping
- concurrency with bounded goroutine lifetimes
- deterministic-enough behavior for tests

## Assumptions (Learning Contract)

- City has `N=20` intersections with IDs `0..N-1`.
- Roads are directed edges in a weighted adjacency list.
- Each road has:
  - `distance` (km)
  - `speedLimit` (km/h)
  - `congestion` in `1..10`
- Route cost is **travel time in minutes**:
  - `base = distance/speedLimit*60`
  - `mult = 1 + (congestion-1)/10`
  - `weight = base*mult`
- Traffic signals per intersection cycle through a circular queue:
  - `[Green-NS, Yellow-NS, Green-EW, Yellow-EW]`
- Left turns are permissive during the owning direction's green (no protected turn phases).
- Congestion evolves from a simple baseline + occupancy model (clamped to `1..10`).
- Sliding window history is last 600 ticks (10 minutes at 1s tick).
- Emergency dispatch:
  - computes the fastest route
  - preempts signals on that route to green for a bounded duration
  - normal vehicles avoid the emergency cor
  ridor via a heavy routing penalty

## Data Structure Mapping

- Intersections: adjacency list graph (`map[int][]Edge`)
- Traffic Signals: circular phase queue per intersection
- Vehicles: hash map (`plate -> vehicle`)
- Emergency Routes: priority queue in Dijkstra
- Congestion History: ring buffer (sliding window)

## Running

Server (REST + pprof):

```bash
go run ./cmd/server -addr :8080 -tick 1s
```

CLI (scenario-driven):

```bash
go run ./cmd/cli -scenario scenarios/01_basic.json
go run ./cmd/cli -scenario scenarios/02_emergency_preempt.json
go run ./cmd/cli -scenario scenarios/03_grid.json
```

The CLI prints a fixed ASCII map once then one-line tick updates until the vehicle arrives or `-max-ticks` is reached. The `-scenario` flag is required. Scenarios are single-vehicle JSON files under `scenarios/`.

### Per-Tick Output Format

Each line shows the vehicle's state at that simulation tick:

**While travelling on a road (`on_road`):**
```
tick=9     CAR-01   normal    on_road     from=2   to=3   rem=3    eta=4.4min  emergency=on(until=tick 65)
```
| Field | Meaning |
|-------|---------|
| `tick=9` | Simulation tick number |
| `CAR-01` | Vehicle plate |
| `normal` | Vehicle type (`normal` / `emergency`) |
| `on_road` | State: travelling between intersections |
| `from=2` | Intersection departed from |
| `to=3` | Intersection heading to |
| `rem=3` | Remaining ticks before arriving at the next intersection |
| `eta=4.4min` | ETA computed once by Dijkstra at registration (simulated minutes) |
| `emergency=on(until=tick 65)` | Emergency corridor active until tick 65 |

**While waiting at a red signal (`waiting_signal` / `at_intersection`):**
```
tick=4     CAR-01   normal    waiting_signal at=0   dest=3   signal=Green-NS   rem=9   q(ns=0 ,ew=1 ) preempt=false  eta=4.4min  emergency=off
```
| Field | Meaning |
|-------|---------|
| `at=0` | Current intersection |
| `dest=3` | Final destination |
| `signal=Green-NS` | Traffic signal phase: cycles `Green-NS` → `Yellow-NS` → `Green-EW` → `Yellow-EW` |
| `rem=9` | Remaining ticks in the current signal phase |
| `q(ns=0, ew=1)` | Queued vehicles waiting for each direction |
| `preempt=false` | `true` = signal forced green by emergency corridor |
| `eta=4.4min` | Dijkstra ETA |

**When arrived:**
```
tick=12    CAR-01   normal    arrived     at=3   dest=3   actual=12/4.4min  eta=4.4min  emergency=off
```
`actual=12/4.4min` = took 12 simulation ticks to complete a 4.4-minute journey.

### How Travel Time is Calculated

### How Travel Time is Calculated

Each road's travel time (in minutes):

```
base_time = distance_km / speed_kmph * 60
multiplier = 1 + (congestion - 1) / 10
weight_minutes = base_time * multiplier
```

The weight is converted to ticks using `ceil(weight_minutes)` with a minimum of 1 tick.

Example: 1km road at 30km/h with congestion=2
- `base_time = 1/30 * 60 = 2.0 minutes`
- `multiplier = 1 + (2-1)/10 = 1.1`
- `weight = 2.0 * 1.1 = 2.2 minutes → 3 ticks`

For 7 hops (03_grid: 0→4→8→12→13→14→15→19): `7 × 2.2 min = 15.4 minutes`

## API

- `GET /healthz`
- `POST /vehicles`
- `GET /route?from=0&to=19`
- `POST /emergency`
- `GET /congestion`
- `GET /signals/{id}`
- `GET /stats`
- `GET /debug/pprof/*`

Example:

```bash
curl -sS localhost:8080/healthz

curl -sS -X POST localhost:8080/vehicles \
  -H 'Content-Type: application/json' \
  -d '{"plate":"KA01AB1234","from":0,"to":19,"type":"normal"}'

curl -sS 'localhost:8080/route?from=0&to=19'
```

## Scenario JSON Format

```json
{
  "name": "01_basic",
  "tickMs": 50,
  "intersections": [
    {"id": 0, "x": 0, "y": 0},
    {"id": 1, "x": 1, "y": 0}
  ],
  "roads": [
    {"id": 1, "from": 0, "to": 1, "distanceKm": 1.0, "speedKmph": 30, "congestion": 2},
    {"id": 101, "from": 1, "to": 0, "distanceKm": 1.0, "speedKmph": 30, "congestion": 2}
  ],
  "vehicle": {"plate": "CAR-01", "type": "normal", "from": 0, "to": 1},
  "events": [
    {"tick": 10, "type": "emergency", "plate": "AMB-01", "from": 0, "to": 1}
  ]
}
```

Fields:
- `tickMs` (required): interval between simulation ticks in milliseconds
- `intersections` (required): array with `id`, `x`, `y` (ids must be contiguous 0..N-1)
- `roads` (required): directed roads with `id`, `from`, `to`, `distanceKm`, `speedKmph`, `congestion` (1..10)
- `vehicle` (required): exactly one vehicle with `plate`, `type` ("normal"|"emergency"), `from`, `to`
- `events` (optional): scheduled events, currently only `"type": "emergency"`

Available scenarios:
- [`scenarios/01_basic.json`](scenarios/01_basic.json) — linear 4-node chain, normal vehicle 0→3
- [`scenarios/02_emergency_preempt.json`](scenarios/02_emergency_preempt.json) — 2×2 grid (4 intersections), vehicle must wait for EW green, emergency at tick 5 preempts signals — observe `preempt=false` → `preempt=true` transition
- [`scenarios/03_grid.json`](scenarios/03_grid.json) — 5×4 grid (20 intersections), normal vehicle 0→19

## Testing

```bash
go test ./...
go test -race ./...
```

Essential end-to-end test cases:
- **HTTP API** (`internal/httpapi/router_test.go`):
  - Health check: `GET /healthz` returns 200.
  - Register + route + simulate arrival:
    - `POST /vehicles` registers a vehicle
    - simulation clock is stepped
    - `GET /stats` eventually reports `VehiclesArrived >= 1`
    - `GET /route` returns a valid path
  - Emergency dispatch preempts signals:
    - `POST /emergency`
    - step simulation clock
    - `GET /signals/0` reports `preempt=true` and a green phase
- **Scenario + CLI** (`internal/scenario/scenario_test.go`, `internal/engine/engine_test.go`):
  - Scenario load and validation (missing fields, bad types, non-contiguous IDs)
  - Vehicle status read API (register, step, assert status fields)
  - Single-vehicle arrival on a chain graph (manual clock steps until `VehicleStateArrived`)
  - Emergency dispatch preempts signals (normal vehicle active, emergency dispatched, signal at intersection 0 reports `preempt=true`)

## Profiling

CPU profile via pprof endpoint:

```bash
go run ./cmd/server
go tool pprof -top http://localhost:8080/debug/pprof/profile?seconds=10
```

## Original Problem Statement

1. CITY MODEL:

- 20 intersections connected by roads (weighted graph)
- Each road: distance, current congestion level (1-10), speed limit
- Represent as adjacency list: map[IntersectionID][]Road

2. FEATURES:
   a) Route Finding:

- Dijkstra's shortest path (implement with your priority queue)
- Input: source intersection, destination intersection
- Output: shortest path, estimated travel time
- Factor in current congestion (dynamic edge weights)
  b) Traffic Signal Management:
- Each intersection has a signal queue (circular): [Green-NS,
  Yellow-NS, Green-EW, Yellow-EW]
- Goroutine per intersection cycling through signals
- Adaptive timing: longer green for congested direction
  c) Vehicle Tracking:
- Register vehicles (HashMap by plate)
- Track current intersection, destination, route
- Move vehicles along routes (simulate with goroutines)
  d) Emergency Vehicle Priority:
- Emergency vehicle enters system → calculate fastest route
  (priority queue)
- Set all signals on route to green (preemption)
- Other vehicles: recalculate routes avoiding emergency path
  e) Congestion Analysis:
- Track congestion per road over time (dynamic array)
- Sliding window average (last 10 minutes)
- Identify top 5 most congested roads
- Suggest alternative routes for congested paths
  f) REST API:
  POST/vehicles→ Register vehicle
  GET/route?from=A&to=B→ Calculate best routePOST/emergency→ Emergency vehicle dispatch
  GET/congestion→ Current congestion heatmap data
  GET/signals/:id→ Signal status at intersection
  GET/stats→ System-wide statistics

3. CONCURRENCY:

- Signal goroutines: one per intersection
- Vehicle goroutines: simulate movement
- Congestion updater: periodic goroutine
- API server: handling requests
- All with proper synchronization (mutex for graph edge weights)
- Graceful shutdown: stop all goroutines cleanly

4. OUTPUT:

- CLI mode: text-based city map showing congestion levels
- API mode: JSON responses for all endpoints
- Logging: all vehicle movements, signal changes, route calculations

5. TESTING:

- Graph operations: shortest path correctness
- Signal cycling: verify correct phase order
- Emergency preemption: verify route clearance
- Concurrent vehicle movement: race detector pass
- Load test: 100 vehicles, 20 intersections

6. PROFILING (mandatory — full profiling suite):

- CPU profile: where does Dijkstra spend the most time? Is it the heap
  or the graph traversal?
- Memory profile: per-vehicle and per-intersection memory footprint
- Goroutine profile: verify all goroutines shut down cleanly (signals,
  vehicles, updater)
- Block profile: are goroutines blocking on channels? Where?
- Benchmark: Dijkstra at 20 vs 100 vs 500 intersections → plot scaling
  curve
- Generate flame graphs for CPU and memory → include in final
  presentation
- Compare: `sync.Mutex` vs `sync.RWMutex` for graph edge weight updates
  under load
