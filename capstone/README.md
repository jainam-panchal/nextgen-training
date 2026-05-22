## Problem Statement

: Smart City Traffic Management
System
Full-day capstone project. Combines ALL Go concepts and ALL data structures.
📦 DS Application Map:
Component
Data Structure
Justification
IntersectionsGraph (adjacency list)Road network topology
Traffic SignalsQueue (circular, per intersection)Cycle through signal phases
VehiclesHashMap (plate → Vehicle)O(1) lookup by ID
Emergency RoutesPriority Queue (min-heap by travel time)Always shortest/fastest route
Congestion HistoryDynamic Array (time-series)Append-only, sliding window
Route SuggestionsTree (decision tree based on conditions)Multi-factor routing decisions
Shortest PathsGraph + Dijkstra (uses priority queue)Optimal route calculation
🏗️ Build Exercise:Build a smart city traffic management system:

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
