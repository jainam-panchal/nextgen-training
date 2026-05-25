# Engine Flow — Tick by Tick

## The Big Picture

The engine runs a loop: for each tick, every goroutine (signals, vehicles, congestion) gets woken up and does its work. The CLI or server drives ticks forward by reading from the clock. Here is the flow for **one tick**:

```
CLI/Server reads next tick from clock
         │
         ▼
Tick driver broadcasts tickN via sync.Cond
         │
         ├──► Intersection goroutines (each independently):
         │     1. intersection.step()
         │     2. markIntersectionProcessed(id, tickN)
         │
         ├──► Vehicle goroutines (each independently):
         │     1. vehicle.step()
         │     2. markVehicleProcessed(plate, tickN)
         │
         └──► Congestion updater:
               1. clear expired emergency corridor
               2. update all road congestion
               3. markCongestionProcessed(tickN)
         │
         ▼
CLI waits for ALL processing via WaitAllProcessed(tickN)
CLI reads and prints VehicleStatus
```

## Line-by-Line: engine.Start()

```go
func (e *Engine) Start(ctx context.Context) {
    ctx, cancel := context.WithCancel(ctx)
    e.runCtx = ctx
    e.stop = cancel
```

Creates a **derived context**. When `Stop()` is called, ALL goroutines see `ctx.Done()` and exit. This is the only shutdown mechanism.

```go
    // Tick driver
    e.wg.Add(1)
    go func() {
        defer e.wg.Done()
        for {
            select {
            case <-ctx.Done():
                e.tickMu.Lock()
                e.tickClosed = true
                e.tickCond.Broadcast()
                e.tickMu.Unlock()
                // ... also broadcast on procCond to unblock WaitAllProcessed
                return
            case tk, ok := <-e.clock.C():
                if !ok {
                    // clock channel closed = shutdown
                    ...
                    return
                }
                e.tickMu.Lock()
                e.tickN = tk.N          // store latest tick
                e.tickCond.Broadcast()  // wake ALL waiters
                e.tickMu.Unlock()
            }
        }
    }()
```

The **tick driver** is the heartbeat. It sits on `clock.C()`, and every time a tick arrives:
1. Stores `tickN` under lock
2. Calls `tickCond.Broadcast()` to wake every goroutine waiting for this tick

When the clock channel closes (ctx cancelled or clock stopped), it sets `tickClosed=true` and broadcasts so all waiters exit.

```go
    // One goroutine per intersection
    for _, inter := range e.signals {
        e.wg.Add(1)
        go func(in *intersection) {
            defer e.wg.Done()
            in.run(ctx, e)
        }(inter)
    }
```

Each intersection gets its own goroutine. Inside `in.run()`:

```go
func (i *intersection) run(ctx context.Context, e *Engine) {
    var last int64
    for {
        n, ok := e.waitTick(last)
        if !ok { return }
        last = n
        if ctx.Err() != nil { return }
        i.step()                                  // <-- advance signal phase
        e.markIntersectionProcessed(i.id, n)       // <-- mark done for barrier
    }
}
```

The pattern is identical for every goroutine:
1. `waitTick(last)` — blocks until tick `last+1` is available
2. Do work
3. `mark*Processed` — tell the barrier that this goroutine finished tick N

```go
    // Congestion updater
    e.wg.Add(1)
    go func() {
        defer e.wg.Done()
        var last int64
        for {
            n, ok := e.waitTick(last)
            if !ok { return }
            last = n
            e.updateCongestion(sim.Tick{N: n})
        }
    }()
}
```

Same pattern. `updateCongestion` runs every tick.

## Line-by-Line: waitTick()

```go
func (e *Engine) waitTick(last int64) (n int64, ok bool) {
    e.tickMu.Lock()
    defer e.tickMu.Unlock()
    for !e.tickClosed && e.tickN < last+1 {
        e.tickCond.Wait()
    }
    if e.tickClosed { return 0, false }
    return last + 1, true
}
```

This is the core synchronization primitive:

- `last` = the last tick this goroutine processed
- It needs tick `last+1`
- If `tickN < last+1`, it waits on `tickCond`
- The tick driver calls `tickCond.Broadcast()` when `tickN` advances
- A goroutine could wake up for a different reason (`tickClosed`), so it loops

Key: `Wait()` **atomically unlocks** `tickMu` and blocks. When woken, it **re-acquires** `tickMu`. This prevents lost wakeups.

## Line-by-Line: intersection.step()

```go
func (i *intersection) step() {
    i.mu.Lock()
    defer i.mu.Unlock()

    if i.preempt {
        i.phase = i.preemptPhase
        i.preemptRemain--
        if i.preemptRemain <= 0 { i.preempt = false }
        return
    }

    i.remaining--
    if i.remaining > 0 { return }

    // Advance circular phase queue
    switch i.phase {
    case PhaseGreenNS:
        i.phase = PhaseYellowNS
        i.remaining = i.cfg.YellowTicks
    case PhaseYellowNS:
        i.phase = PhaseGreenEW
        i.remaining = i.adaptiveGreenTicks("EW")
    case PhaseGreenEW:
        i.phase = PhaseYellowEW
        i.remaining = i.cfg.YellowTicks
    case PhaseYellowEW:
        i.phase = PhaseGreenNS
        i.remaining = i.adaptiveGreenTicks("NS")
    }
}
```

Every tick:
1. **If preempted**: lock signal to the preempt phase, decrement preempt timer.
2. **Otherwise**: decrement remaining ticks for current phase.
3. **If phase expired**: move to next phase in the cycle.

The phase cycle is: `Green-NS → Yellow-NS → Green-EW → Yellow-EW → Green-NS → ...`

**Adaptive green time**: `adaptiveGreenTicks(group)` uses queue length ratios to extend green for the congested direction, clamped to `[MinGreenTicks, MaxGreenTicks]`.

## Line-by-Line: vehicle.step()

```go
func (v *vehicle) step(ctx context.Context, e *Engine, tk sim.Tick) {
    v.mu.Lock()
    defer v.mu.Unlock()

    if v.state == VehicleStateArrived { return }

    // Phase 1: Advance if on road
    if v.state == VehicleStateOnRoad {
        v.remainingRoad--
        if v.remainingRoad > 0 { return }   // still travelling
        // Arrived at next intersection
        v.at = v.nextIntersection
        v.state = VehicleStateAtIntersection
    }

    // Phase 2: Check if at destination
    if v.at == v.to {
        v.state = VehicleStateArrived
        return
    }

    // Phase 3: Re-route periodically
    if tk.N - v.lastRouteTick >= ... {
        path, _, _ := e.route(v.at, v.to, v.kind)
        v.path = path; v.idx = 0
    }

    // Phase 4: Get next hop from route
    from := v.path[v.idx]
    to := v.path[v.idx+1]

    // Phase 5: Check signal
    if phaseAllows(phase, movementDirection(from, to)) {
        // GREEN — enter road
        v.state = VehicleStateOnRoad
        v.remainingRoad = travelTicks
        v.nextIntersection = to
        v.idx++
    } else {
        // RED — queue at intersection
        v.state = VehicleStateWaitingSignal
    }
}
```

Each vehicle tick:
1. If currently **on a road**, decrement the travel timer. When it hits 0, the vehicle arrives at the next intersection.
2. If at destination (`v.at == v.to`), mark arrived.
3. **Re-route** every `ReRouteEveryTicks` ticks — recalculates path with current congestion.
4. Look at the **next hop** in the path.
5. Check the **signal phase** at the current intersection:
   - If green in the right direction → enter the road (set travel timer, change state to `on_road`)
   - If red → queue up, change state to `waiting_signal`

## Line-by-Line: emergency corridor

```go
func (e *Engine) DispatchEmergency(ctx, req) (EmergencyResponse, error) {
    path, eta, _ := e.route(req.From, req.To, VehicleTypeEmergency)
    e.setEmergencyCorridor(path)
    e.RegisterVehicle(ctx, ...)  // register the emergency vehicle
    return ...
}

func (e *Engine) setEmergencyCorridor(path []int) {
    // 1. Store expiry tick
    e.emergencyUntilTick = now + PreemptionTicks

    // 2. Mark all roads on path as corridor
    for i := 0; i < len(path)-1; i++ {
        e.roadOnCorr[road.ID] = struct{}{}
    }

    // 3. Preempt each intersection on the path
    for i := 0; i < len(path)-1; i++ {
        inter := e.signals[path[i]]
        if movementDirection(path[i], path[i+1]) == "NS" {
            inter.setPreempt(PhaseGreenNS, PreemptionTicks)
        } else {
            inter.setPreempt(PhaseGreenEW, PreemptionTicks)
        }
    }
}
```

When emergency is dispatched:
1. Compute fastest route (penalty-free, since `kind == VehicleTypeEmergency`).
2. Mark all roads on that route as **corridor** (stored in `roadOnCorr` map).
3. For each intersection on the route, force the signal to green in the emergency vehicle's direction for `PreemptionTicks` ticks.
4. Register the emergency vehicle (which will traverse the route with all-green signals).

**For normal vehicles**: the `graphView.Neighbors()` method checks `roadOnCorr`. If a normal vehicle is routing and a road is on the corridor, it gets congestion=10 and distance×50 — effectively routing around the emergency.

**Corridor expiry**: The congestion updater clears `roadOnCorr` and sets `emergencyUntilTick=0` when the current tick exceeds the expiry.

## Line-by-Line: WaitAllProcessed (the barrier)

```go
func (e *Engine) WaitAllProcessed(tickN int64) bool {
    e.procMu.Lock()
    defer e.procMu.Unlock()
    for !e.tickClosed && !e.allProcessedLocked(tickN) {
        e.procCond.Wait()          // blocks until all goroutines catch up
    }
    return e.allProcessedLocked(tickN)
}

func (e *Engine) allProcessedLocked(tickN int64) bool {
    for id := 0; id < e.cfg.NumIntersections; id++ {
        if e.interLast[id] < tickN { return false }
    }
    for _, v := range e.vehicleLast {
        if v < tickN { return false }
    }
    return e.congLast >= tickN
}
```

This is used by the CLI to ensure it reads **consistent state**. Without this barrier:
- Signal might have processed tick 10, but the vehicle might still be on tick 9
- The CLI would read stale vehicle position

Each goroutine calls `mark*Processed(tickN)` after finishing tick N. The barrier wakes up, checks every goroutine, and returns when all have caught up.

## Line-by-Line: Congestion Update

```go
func (e *Engine) updateCongestion(tk sim.Tick) {
    e.roadMu.Lock()
    defer e.roadMu.Unlock()

    // Clear expired corridor
    if e.emergencyUntilTick > 0 && tk.N > e.emergencyUntilTick {
        e.roadOnCorr = make(map[int]struct{})
        e.emergencyUntilTick = 0
    }

    for _, rs := range e.roadByID {
        baseline := 2
        occImpact := rs.occupancy / 5
        v := clampCongestion(baseline + occImpact)
        rs.congestion = v
        rs.history.Append(v)     // sliding window for avg
    }
}
```

Every tick:
1. Check if the emergency corridor has expired.
2. For every road: `congestion = 2 + (occupancy / 5)`, clamped to `1..10`.
   - `baseline=2` represents normal background traffic
   - `occupancy` = number of vehicles currently on this road
   - Each 5 vehicles adds 1 congestion level
3. Append to ring buffer (600 entries = 10 min window at 1s/tick).

## Line-by-Line: Dijkstra Routing

```go
func ShortestPath(g model.GraphReader, src, dst int) ([]int, float64) {
    // Initialize all distances to infinity
    // Set dist[src] = 0
    // Push src onto priority queue with priority 0
    for {
        u, d, ok := pq.PopItem()    // get node with smallest distance
        if !ok { break }
        if d > dist[u] { continue } // skip stale entry
        if u == dst { break }       // reached destination

        for _, e := range g.Neighbors(u) {
            v := e.To
            alt := dist[u] + WeightMinutes(e.Road)  // cost in minutes
            if alt < dist[v] {
                dist[v] = alt
                prev[v] = u
                pq.PushItem(v, alt)  // may push duplicate; stale check handles it
            }
        }
    }
    // Reconstruct path by following prev[] backwards from dst
}
```

Standard Dijkstra with a min-heap priority queue. The cost function `WeightMinutes` uses the dynamic road data from `graphView`, which reads current congestion from the engine under `roadMu.RLock()`.

Priority queue is a generic `container/heap` wrapper (supports any type `T`).
