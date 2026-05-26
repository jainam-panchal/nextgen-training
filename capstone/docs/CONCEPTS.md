# Key Concepts Explained

## 1. Simulation Tick

The entire simulation runs on **discrete ticks**. Think of a tick as one turn in a board game — every piece moves or acts exactly once per turn.

**What happens in 1 tick:**
- Every intersection's signal phase updates (or stays the same)
- Every vehicle on a road advances 1 tick closer to its destination
- Every vehicle at a red light waits (or gets to go if the light turned green)
- Road congestion is recalculated

**Wall clock vs tick:** A tick has a real-world duration (`TickDuration`), but all internal logic uses tick numbers. The CLI uses 50ms ticks (runs fast), the server uses 1s ticks (human-readable). The simulation produces the same result regardless of tick speed.

**Why not real time?** Real-time simulation is non-deterministic — goroutines race, timers drift. Integer ticks make testing trivial (`ManualClock.Step()` = advance 1 tick), race-free, and deterministic.

## 2. The Clock Abstraction

```go
type Clock interface {
    C() <-chan Tick
}
```

Two implementations:

- **`RealClock`**: Wraps `time.Ticker`. Emits ticks at a fixed interval. Used by CLI and server.
- **`ManualClock`**: Channel you push ticks into via `.Step()`. Used by tests. Each `.Step()` = 1 tick.

The engine does not know or care which clock it's using. It just reads from `clock.C()`.

**Test trick:** With `ManualClock`, you can step one tick, call `WaitAllProcessed(1)`, and assert exact state. No `time.Sleep` needed.

## 3. sync.Cond — The Tick Broadcast Pattern

`sync.Cond` is Go's condition variable. It lets one goroutine wake up many others at once.

**The problem:** If we used a channel for ticks, only ONE goroutine would receive each tick (channels are point-to-point). We need N goroutines (signals, vehicles, congestion) to ALL process the same tick.

**The solution:** The tick driver broadcasts on `tickCond`. All waiting goroutines wake up, read `tickN`, and process.

```
tickCond.Wait() does:
  1. Atomically unlock tickMu
  2. Block until Broadcast/Signal
  3. Re-lock tickMu
```

This guarantees there's no race between "check tickN" and "wait for next tick."

## 4. WaitVehicleProcessed — Per-Vehicle Sync (Replaces Barrier)

The old `WaitAllProcessed` barrier tracked every goroutine's tick progress via `mark*Processed()` calls and a shared `sync.Cond`. This was removed.

Instead, each vehicle atomically stores its last processed tick after every `step()`:

```go
v.step(e)
atomic.StoreInt64(&v.processedTick, last)
```

Tests call `WaitVehicleProcessed(plate, tickN)` which busy-waits until that specific vehicle reaches the target tick. This is lighter and sufficient for test determinism — the CLI doesn't need a barrier because it only reads one vehicle's state.

## 5. Dijkstra — Shortest Path with Dynamic Weights

Standard Dijkstra using a min-heap priority queue. The key twist: **road weights change every tick** based on congestion.

```go
// Fresh weight read every time Dijkstra runs:
alt = dist[u] + WeightMinutes(e.Road)

// WeightMinutes:
func WeightMinutes(r *model.Road) float64 {
    base := r.Distance / r.SpeedLimit     // hours
    mult := 1.0 + float64(r.Congestion-1)/10.0
    return base * mult * 60.0             // minutes
}
```

**Example:** 1km road, 30km/h, congestion=2
- `base = 1/30 = 0.0333 hours = 2.0 minutes`
- `mult = 1 + (2-1)/10 = 1.1`
- `weight = 2.0 × 1.1 = 2.2 minutes`

**Conversion to ticks:** `ceil(weight)` with minimum 1 tick. So 2.2 min → 3 ticks.

**graphView:** The engine implements `GraphReader` via `graphView`. When Dijkstra calls `Neighbors(node)`, `graphView` reads the **current** congestion from the engine's `roadByID` map (under `roadMu.RLock()`). Every routing query gets live congestion data.

## 6. Traffic Signal Phases

Each intersection cycles through 4 phases:

```
Green-NS  →  Yellow-NS  →  Green-EW  →  Yellow-EW  →  Green-NS → ...
  (NS moves)   (warning)    (EW moves)   (warning)
```

**Green duration:** Adaptive based on queue lengths. If 10 cars are waiting NS and 2 are waiting EW, the NS green gets longer.

```
ratio = queue_ns / (queue_ns + queue_ew)
green_ticks = minGreen + ratio × (maxGreen - minGreen)
```

Clamped to `[MinGreenTicks, MaxGreenTicks]` (default 10..30).

**Default config:** MinGreen=10 ticks, MaxGreen=30 ticks, Yellow=3 ticks.

## 7. movementDirection — NS vs EW Classification

```go
func movementDirection(from, to int) string {
    if to-from == 1 || from-to == 1 {
        return "NS"
    }
    return "EW"
}
```

This simple rule works because of **column-major grid IDs**:

```
Column 0: IDs 0, 1, 2, 3    (adjacent = NS)
Column 1: IDs 4, 5, 6, 7
...
Adjacent across columns: 0→4 (diff=4) → EW
```

In the chain topology (01_basic: 0→1→2→3), adjacent IDs mean: the map is a single row, so "NS" really means "east" along the row. This is a learning-PoC simplification; a real system would use actual coordinates.

## 8. Vehicle State Machine

```
                      ┌─────────────────────────────┐
                      │                             │
                      v                             │
  at_intersection ──► waiting_signal ──► on_road ──┘
       │                                  │
       │                                  │
       └──────────────────────────────────┘
                    │
                    v
                arrived
```

**at_intersection:** Vehicle just arrived at a new intersection. Checks signal. If green in the right direction → `on_road`. If red → `waiting_signal`.

**waiting_signal:** Waiting at red light. On each tick, the vehicle checks the signal phase. When green → `on_road`.

**on_road:** Travelling along a road. Each tick decrements `remainingRoad`. When it hits 0 → `at_intersection`.

**arrived:** Destination reached. Terminal state.

## 9. Emergency Preemption

When `DispatchEmergency` is called:

1. **Route calculation:** Emergency route computed with NO penalty (normal route penalty is disabled for `VehicleTypeEmergency`).

2. **Corridor marking:** All roads on the emergency route are marked in `roadOnCorr` map. Normal vehicles routing queries see these roads with congestion=10 and distance×50 (massive penalty) — they route around.

3. **Signal preemption:** Every intersection on the route gets `setPreempt(phase, ticks)` where `phase` is the direction the emergency vehicle needs (NS or EW). The intersection stays in that phase, ignoring its normal cycle, for `PreemptionTicks` (default 60 ticks).

4. **Corridor expiry:** After `PreemptionTicks`, the corridor is cleared by the congestion updater. Signals resume normal cycling. Normal vehicle routing penalties are removed.

**Sequence:** (from `02_emergency_preempt.json`)
```
t=1   CAR-01 waits at intersection 0 for Green-NS
      (normal traffic)
t=5   AMB-01 emergency dispatched from 0→3
      → route [0 2 3], preempting 2 intersections
      → intersection 0 forced Green-EW (AMB-01's direction)
t=6   CAR-01 also has Green-EW (needs NS to go 0→2... 
      actually 0→2 is EW in the grid)
t=7   CAR-01 enters road 0→2
      ...
t=12  CAR-01 arrives
```

Note: In the grid, `movementDirection(0,2)` = EW (diff=2, not 1). The emergency also goes 0→2, so both get the EW green. If the normal vehicle needed the opposite direction, it would wait until the corridor expires.

## 10. Congestion Model

```go
congestion = clamp(2 + occupancy / 5, 1, 10)
```

- **Baseline:** 2 (light traffic)
- **Occupancy:** Number of vehicles currently on that road
- **Impact:** Each 5 vehicles adds 1 congestion level
- **Clamp:** Always 1..10

Updated every tick by the congestion updater goroutine.

A **sliding window** of the last `HistoryWindow` (default 600) congestion values is stored in a `RingBuffer`. The API's `GET /congestion` returns both current and average congestion.

## 11. Scenario JSON Format

Two road formats:

**Verbose (object array):**
```json
"roads": [
    {"id": 1, "from": 0, "to": 1, "distanceKm": 1.0, "speedKmph": 30, "congestion": 2}
]
```

**Compact (array of arrays):**
```json
"roads": [[0, 1, 1.0, 30, 2]]
```
`[from, to, distanceKm, speedKmph, congestion]`

**Grid shorthand:** Generate N×M grid with uniform roads.
```json
"grid": {"cols": 5, "rows": 4, "cellDist": 1.0, "speed": 30, "congestion": 2}
```

**Bidirectional flag:** When `true`, every road entry automatically gets a reverse counterpart. Works with both verbose and compact formats.

## 12. Thread Safety Map

Every shared data structure has its own mutex strategy:

| Data | Protections | Why |
|------|-------------|-----|
| `vehicles` (map) | `engine.mu` (RWMutex) | Registered/read by API calls and CLI |
| `roadByID` (congestion, occupancy, corridor) | `engine.roadMu` (RWMutex) | Updated by congestion updater, read by graphView |
| `intersection` (phase, queue, preempt) | Per-intersection `mu` (Mutex) | Updated by intersection goroutine, read by vehicle goroutines and API |
| `vehicle` (state, position, path) | Per-vehicle `mu` (Mutex) | Updated by own goroutine, read by VehicleStatus API |
| `tickN` | `engine.tickMu` (Mutex) | Written by tick driver, read by waitTick |

No two goroutines write to the same position without synchronization. The `graphView.Neighbors()` snapshot pattern (copy Road struct under lock, return copy) ensures Dijkstra never reads a partially-updated road.

## 13. Single-Condvar Design

The engine originally had two `sync.Cond` instances: `tickCond` (tick broadcast) and `procCond` (progress barrier). The `procCond` was removed because:

- It added ~80 lines of machinery (`procMu`, `procCond`, `interLast`, `vehicleLast`, `congLast`, `WaitAllProcessed`, `mark*Processed`, `allProcessedLocked`)
- The CLI doesn't need a full barrier — it only reads one vehicle's state, which is consistent after `waitTick` returns
- Tests use the lighter `WaitVehicleProcessed` per-vehicle atomic counter instead

Now the engine has exactly one `sync.Cond` (`tickCond`), owned solely by the tick driver. All goroutines (intersections, vehicles, congestion updater) are symmetric waiters.

## 14. Goroutine Safe Shutdown

```
engine.Stop() →
  ctx cancelled (all goroutines see ctx.Done())
  tickClosed = true, cond.Broadcast()
  waitTick() returns ok=false
  Each goroutine: defer wg.Done()
  wg.Wait() in Stop()
```

The `closeOnce` sync ensures `Stop()` is idempotent. Both the tick driver and signal goroutines check `ctx.Err()` and `tickClosed` to exit cleanly.
