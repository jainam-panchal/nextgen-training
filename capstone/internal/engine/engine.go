package engine

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/dijkstra"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/ds"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/sim"
)

// Engine is the core simulation engine. It owns the road graph, traffic signals,
// vehicles, congestion state, and manages the tick-driven concurrent execution.
//
// Goroutine lifecycle (managed by Start/Stop):
//   - 1 tick-driver goroutine: reads clock ticks and broadcasts via tickCond
//   - N intersection goroutines: one per node, step signal phases each tick
//   - 1 congestion-updater goroutine: recomputes road congestion each tick
//   - M vehicle goroutines: one per registered vehicle, step state machine each tick
type Engine struct {
	cfg   Config
	graph *model.Graph

	mu       sync.RWMutex
	signals  map[int]*intersection // one signal controller per intersection node
	vehicles map[string]*vehicle   // active vehicles indexed by plate

	roadMu             sync.RWMutex // guards roadByID, roadOnCorr, emergencyUntilTick
	roadByID           map[int]*roadState
	roadOnCorr         map[int]struct{} // road IDs that are part of the active emergency corridor
	emergencyUntilTick int64            // tick at which the emergency corridor expires

	clock sim.Clock // source of ticks (RealClock or ManualClock)

	runCtx context.Context // cancelled by Stop()
	stop   context.CancelFunc

	wg        sync.WaitGroup // tracks all goroutines; Stop() waits here
	closeOnce sync.Once

	tickMu     sync.Mutex // shared with tickCond
	tickCond   *sync.Cond // broadcasts each tick to wake all goroutines
	tickN      int64      // current tick number (set by tick driver)
	tickClosed bool       // true after Stop() to unblock waiters
}

// movementDirection classifies a hop as "NS" (north-south) or "EW" (east-west)
// based on intersection IDs. Works because grid scenarios use column-major numbering
// where adjacent IDs are vertical neighbours.
func movementDirection(from, to int) string {
	if to-from == 1 || from-to == 1 {
		return "NS"
	}
	return "EW"
}

// New creates an Engine. Validates config, pre-allocates intersections, indexes roads.
func New(cfg Config, g *model.Graph, clock sim.Clock) (*Engine, error) {
	if cfg.NumIntersections <= 0 {
		return nil, errors.New("NumIntersections must be > 0")
	}
	if g == nil {
		return nil, errors.New("graph is nil")
	}
	if clock == nil {
		return nil, errors.New("clock is nil")
	}
	if cfg.HistoryWindow <= 0 {
		return nil, errors.New("HistoryWindow must be > 0")
	}

	e := &Engine{
		cfg:        cfg,
		graph:      g,
		signals:    make(map[int]*intersection, cfg.NumIntersections),
		vehicles:   make(map[string]*vehicle),
		roadByID:   make(map[int]*roadState),
		roadOnCorr: make(map[int]struct{}),
		clock:      clock,
	}
	e.tickCond = sync.NewCond(&e.tickMu)

	// Create one signal controller per intersection node.
	for i := 0; i < cfg.NumIntersections; i++ {
		e.signals[i] = newIntersection(i, cfg)
	}

	// Index every unique road in the graph into roadByID with initial congestion state.
	for _, edges := range g.Adj {
		for _, edge := range edges {
			r := edge.Road
			if r == nil {
				continue
			}
			if _, ok := e.roadByID[r.ID]; ok {
				continue
			}
			e.roadByID[r.ID] = &roadState{
				road:       *r,
				congestion: clampCongestion(r.Congestion),
				history:    ds.NewRingBuffer[int](cfg.HistoryWindow),
			}
		}
	}
	return e, nil
}

// Start launches the simulation: tick driver, intersection goroutines, congestion updater.
// Blocks the calling goroutine only during setup; all loops run in background goroutines.
func (e *Engine) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	e.runCtx = ctx
	e.stop = cancel

	// Tick-driver goroutine: reads ticks from the clock and broadcasts them
	// to all waiters via tickCond. This is the sole producer of tick events.
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
				return
			case tk, ok := <-e.clock.C():
				if !ok {
					e.tickMu.Lock()
					e.tickClosed = true
					e.tickCond.Broadcast()
					e.tickMu.Unlock()
					return
				}
				e.tickMu.Lock()
				e.tickN = tk.N
				e.tickCond.Broadcast()
				e.tickMu.Unlock()
			}
		}
	}()

	// One goroutine per intersection signal controller.
	for _, inter := range e.signals {
		e.wg.Add(1)
		go func(in *intersection) {
			defer e.wg.Done()
			in.run(ctx, e)
		}(inter)
	}

	// Congestion-updater goroutine: recomputes road congestion every tick.
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		var last int64
		for {
			n, ok := e.waitTick(last)
			if !ok {
				return
			}
			last = n
			e.updateCongestion(sim.Tick{N: n})
		}
	}()
}

// Stop cancels the run context, closes tick broadcast, and waits for all goroutines.
func (e *Engine) Stop() {
	e.closeOnce.Do(func() {
		if e.stop != nil {
			e.stop()
		}
		e.tickMu.Lock()
		e.tickClosed = true
		e.tickCond.Broadcast()
		e.tickMu.Unlock()
		e.wg.Wait()
	})
}

// WaitNextTick blocks until tick last+1 has been broadcast. Used by CLI consumers.
func (e *Engine) WaitNextTick(last int64) (n int64, ok bool) {
	return e.waitTick(last)
}

// waitTick is the core synchronization point. Every goroutine calls this to
// wait for the next tick. The tick driver broadcasts on tickCond after reading
// a new tick from the clock, which wakes all waiters simultaneously.
func (e *Engine) waitTick(last int64) (n int64, ok bool) {
	e.tickMu.Lock()
	defer e.tickMu.Unlock()
	for !e.tickClosed && e.tickN < last+1 {
		e.tickCond.Wait()
	}
	if e.tickClosed {
		return 0, false
	}
	return last + 1, true
}

// updateCongestion runs once per tick. It:
//   - Clears the emergency corridor when it expires
//   - Recomputes each road's congestion from occupancy: base=2 + occupancy/5, clamped [1,10]
//   - Appends the new value to the history ring buffer for rolling averages
func (e *Engine) updateCongestion(tk sim.Tick) {
	e.roadMu.Lock()
	defer e.roadMu.Unlock()

	// Expire the emergency corridor if its time has passed.
	if e.emergencyUntilTick > 0 && tk.N > e.emergencyUntilTick {
		e.roadOnCorr = make(map[int]struct{})
		e.emergencyUntilTick = 0
	}
	for _, rs := range e.roadByID {
		baseline := 2
		occImpact := rs.occupancy / 5
		v := clampCongestion(baseline + occImpact)
		rs.congestion = v
		rs.history.Append(v)
	}
}

// clampCongestion keeps congestion in [1, 10].
func clampCongestion(v int) int {
	if v < 1 {
		return 1
	}
	if v > 10 {
		return 10
	}
	return v
}

// --- Public API ---

// RegisterVehicleRequest is the JSON body for POST /vehicles.
type RegisterVehicleRequest struct {
	Plate string      `json:"plate"`
	From  int         `json:"from"`
	To    int         `json:"to"`
	Type  VehicleType `json:"type"`
}

// VehicleStatus is a snapshot of a vehicle's current state, returned by VehicleStatus().
type VehicleStatus struct {
	Plate              string       `json:"plate"`
	Type               VehicleType  `json:"type"`
	State              VehicleState `json:"state"`
	At                 int          `json:"at"`
	Next               int          `json:"next"`
	To                 int          `json:"to"`
	OnRoadID           int          `json:"onRoadId"`
	RemainingRoadTicks int          `json:"remainingRoadTicks"`
}

// VehicleStatus returns a snapshot of the vehicle's current state.
// Next is computed dynamically from path[idx] (destination of current road)
// or path[idx+1] (next hop) if the vehicle is at an intersection.
func (e *Engine) VehicleStatus(ctx context.Context, plate string) (VehicleStatus, error) {
	if err := ctx.Err(); err != nil {
		return VehicleStatus{}, fmt.Errorf("context: %w", err)
	}
	e.mu.RLock()
	v := e.vehicles[plate]
	e.mu.RUnlock()
	if v == nil {
		return VehicleStatus{}, errors.New("vehicle not found")
	}
	v.mu.Lock()
	st := VehicleStatus{
		Plate:              v.plate,
		Type:               v.kind,
		State:              v.state,
		At:                 v.at,
		To:                 v.to,
		OnRoadID:           v.onRoadID,
		RemainingRoadTicks: v.remainingRoad,
	}
	// Compute Next dynamically from path position.
	if v.state == VehicleStateOnRoad && v.idx < len(v.path) {
		st.Next = v.path[v.idx] // destination of current road traversal
	} else if len(v.path) > v.idx+1 {
		st.Next = v.path[v.idx+1] // next hop
	} else {
		st.Next = v.at // fallback: current position
	}
	v.mu.Unlock()
	return st, nil
}

// WaitVehicleProcessed busy-waits until the vehicle goroutine has processed tickN.
// Used by tests for deterministic synchronization.
func (e *Engine) WaitVehicleProcessed(plate string, tickN int64) {
	e.mu.RLock()
	v := e.vehicles[plate]
	e.mu.RUnlock()
	if v == nil {
		return
	}
	for atomic.LoadInt64(&v.processedTick) < tickN {
	}
}

// RouteResponse contains the computed path and estimated travel time in minutes.
type RouteResponse struct {
	Path       []int   `json:"path"`
	ETAMinutes float64 `json:"etaMinutes"`
}

// RegisterVehicle creates a vehicle, computes a Dijkstra route, stores it,
// and launches its goroutine. Returns the route and ETA.
func (e *Engine) RegisterVehicle(ctx context.Context, req RegisterVehicleRequest) (RouteResponse, error) {
	if err := ctx.Err(); err != nil {
		return RouteResponse{}, fmt.Errorf("context: %w", err)
	}
	if req.Plate == "" {
		return RouteResponse{}, errors.New("plate is required")
	}
	if req.From < 0 || req.From >= e.cfg.NumIntersections {
		return RouteResponse{}, errors.New("from out of range")
	}
	if req.To < 0 || req.To >= e.cfg.NumIntersections {
		return RouteResponse{}, errors.New("to out of range")
	}
	if req.Type == "" {
		req.Type = VehicleTypeNormal
	}
	if req.Type != VehicleTypeNormal && req.Type != VehicleTypeEmergency {
		return RouteResponse{}, errors.New("invalid vehicle type")
	}

	path, eta, err := e.route(req.From, req.To, req.Type)
	if err != nil {
		return RouteResponse{}, err
	}

	e.mu.Lock()
	if _, ok := e.vehicles[req.Plate]; ok {
		e.mu.Unlock()
		return RouteResponse{}, errors.New("vehicle already exists")
	}
	v := &vehicle{
		plate: req.Plate,
		kind:  req.Type,
		to:    req.To,
		state: VehicleStateAtIntersection,
		at:    req.From,
		path:  append([]int(nil), path...),
		idx:   0,
	}
	e.vehicles[req.Plate] = v
	e.mu.Unlock()

	// Launch the vehicle's goroutine.
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		v.run(e.runCtx, e)
	}()

	return RouteResponse{Path: path, ETAMinutes: eta}, nil
}

// Route computes a shortest path between two intersections (normal vehicle routing).
func (e *Engine) Route(ctx context.Context, from, to int) (RouteResponse, error) {
	if err := ctx.Err(); err != nil {
		return RouteResponse{}, fmt.Errorf("context: %w", err)
	}
	if from < 0 || from >= e.cfg.NumIntersections {
		return RouteResponse{}, errors.New("from out of range")
	}
	if to < 0 || to >= e.cfg.NumIntersections {
		return RouteResponse{}, errors.New("to out of range")
	}
	path, eta, err := e.route(from, to, VehicleTypeNormal)
	if err != nil {
		return RouteResponse{}, err
	}
	return RouteResponse{Path: path, ETAMinutes: eta}, nil
}

// DispatchEmergencyRequest is the JSON body for POST /emergency.
type DispatchEmergencyRequest struct {
	Plate string `json:"plate"`
	From  int    `json:"from"`
	To    int    `json:"to"`
}

// EmergencyResponse contains the emergency route and preemption info.
type EmergencyResponse struct {
	RouteResponse
	PreemptionTicks int `json:"preemptionTicks"`
}

// DispatchEmergency creates an emergency corridor along the shortest path:
//   - Routes the emergency vehicle (sees all roads normally)
//   - Marks corridor roads so normal vehicles avoid them
//   - Preempts signal controllers along the path to stay green
//   - Registers the emergency vehicle and starts it
func (e *Engine) DispatchEmergency(ctx context.Context, req DispatchEmergencyRequest) (EmergencyResponse, error) {
	if err := ctx.Err(); err != nil {
		return EmergencyResponse{}, fmt.Errorf("context: %w", err)
	}
	if req.Plate == "" {
		return EmergencyResponse{}, errors.New("plate is required")
	}
	path, eta, err := e.route(req.From, req.To, VehicleTypeEmergency)
	if err != nil {
		return EmergencyResponse{}, err
	}

	e.setEmergencyCorridor(path)

	_, err = e.RegisterVehicle(context.Background(), RegisterVehicleRequest{Plate: req.Plate, From: req.From, To: req.To, Type: VehicleTypeEmergency})
	if err != nil {
		return EmergencyResponse{}, err
	}

	return EmergencyResponse{RouteResponse: RouteResponse{Path: path, ETAMinutes: eta}, PreemptionTicks: e.cfg.PreemptionTicks}, nil
}

// setEmergencyCorridor marks all roads along path as corridor (normal vehicles avoid)
// and preempts signal controllers at each intersection to stay green for the
// emergency vehicle's direction. The corridor expires after PreemptionTicks ticks.
func (e *Engine) setEmergencyCorridor(path []int) {
	if len(path) < 2 {
		return
	}

	e.tickMu.Lock()
	now := e.tickN
	e.tickMu.Unlock()

	e.roadMu.Lock()
	e.roadOnCorr = make(map[int]struct{})
	e.emergencyUntilTick = now + int64(e.cfg.PreemptionTicks)
	e.roadMu.Unlock()

	// Mark all road IDs along the path as corridor.
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		for _, edge := range e.graph.Adj[from] {
			if edge.To == to && edge.Road != nil {
				e.roadMu.Lock()
				e.roadOnCorr[edge.Road.ID] = struct{}{}
				e.roadMu.Unlock()
				break
			}
		}
	}

	// Preempt signal controllers at each intersection along the path.
	for i := 0; i < len(path)-1; i++ {
		interID := path[i]
		next := path[i+1]
		inter := e.signals[interID]
		if inter == nil {
			continue
		}
		if movementDirection(interID, next) == "NS" {
			inter.setPreempt(PhaseGreenNS, e.cfg.PreemptionTicks)
		} else {
			inter.setPreempt(PhaseGreenEW, e.cfg.PreemptionTicks)
		}
	}
}

// SignalStatus describes the state of a traffic signal at one intersection.
type SignalStatus struct {
	IntersectionID int   `json:"intersectionId"`
	Phase          Phase `json:"phase"`
	RemainingTicks int   `json:"remainingTicks"`
	QueueNS        int   `json:"queueNS"`
	QueueEW        int   `json:"queueEW"`
	Preempt        bool  `json:"preempt"`
}

// SignalStatus returns the current state of a traffic signal.
func (e *Engine) SignalStatus(_ context.Context, id int) (SignalStatus, error) {
	if id < 0 || id >= e.cfg.NumIntersections {
		return SignalStatus{}, errors.New("id out of range")
	}
	inter := e.signals[id]
	if inter == nil {
		return SignalStatus{}, errors.New("intersection not found")
	}
	st := inter.status()
	st.IntersectionID = id
	return st, nil
}

// RoadCongestion describes congestion on a single road.
type RoadCongestion struct {
	RoadID           int     `json:"roadId"`
	From             int     `json:"from"`
	To               int     `json:"to"`
	Congestion       int     `json:"congestion"`
	AvgCongestion    float64 `json:"avgCongestion"`
	Occupancy        int     `json:"occupancy"`
	OnEmergencyRoute bool    `json:"onEmergencyRoute"`
}

// CongestionResponse groups all roads (sorted by avg congestion descending) + top 5.
type CongestionResponse struct {
	Roads []RoadCongestion `json:"roads"`
	Top5  []RoadCongestion `json:"top5"`
}

// Congestion returns congestion data for all roads, sorted by average congestion descending.
func (e *Engine) Congestion(_ context.Context) CongestionResponse {
	e.roadMu.RLock()
	defer e.roadMu.RUnlock()
	roads := make([]RoadCongestion, 0, len(e.roadByID))
	for _, rs := range e.roadByID {
		avg := meanInts(rs.history.Items())
		_, on := e.roadOnCorr[rs.road.ID]
		roads = append(roads, RoadCongestion{
			RoadID:           rs.road.ID,
			From:             rs.road.From,
			To:               rs.road.To,
			Congestion:       rs.congestion,
			AvgCongestion:    avg,
			Occupancy:        rs.occupancy,
			OnEmergencyRoute: on,
		})
	}
	sort.Slice(roads, func(i, j int) bool { return roads[i].AvgCongestion > roads[j].AvgCongestion })
	top := 5
	if len(roads) < top {
		top = len(roads)
	}
	return CongestionResponse{Roads: roads, Top5: append([]RoadCongestion(nil), roads[:top]...)}
}

func meanInts(v []int) float64 {
	if len(v) == 0 {
		return 0
	}
	var sum int
	for _, x := range v {
		sum += x
	}
	return float64(sum) / float64(len(v))
}

// Stats groups summary statistics about the simulation.
type Stats struct {
	VehiclesTotal   int `json:"vehiclesTotal"`
	VehiclesArrived int `json:"vehiclesArrived"`
}

// Stats returns summary counts.
func (e *Engine) Stats(_ context.Context) Stats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	arrived := 0
	for _, v := range e.vehicles {
		v.mu.Lock()
		st := v.state
		v.mu.Unlock()
		if st == VehicleStateArrived {
			arrived++
		}
	}
	return Stats{VehiclesTotal: len(e.vehicles), VehiclesArrived: arrived}
}

// route computes a shortest path using Dijkstra. The graphView adapts the road
// network with live congestion and (for normal vehicles) emergency corridor penalties.
func (e *Engine) route(from, to int, kind VehicleType) ([]int, float64, error) {
	gr := &graphView{e: e, avoidEmergency: kind != VehicleTypeEmergency}
	path, w := dijkstra.ShortestPath(gr, from, to)
	if path == nil || math.IsInf(w, 1) {
		return nil, 0, errors.New("no route")
	}
	return path, w, nil
}

// --- graph view and road state ---

// roadState tracks live state for a single road, updated every tick.
type roadState struct {
	road       model.Road          // static road definition (ID, from, to, distance, speed)
	congestion int                 // current congestion level [1,10]
	occupancy  int                 // number of vehicles currently on this road
	history    *ds.RingBuffer[int] // rolling congestion history (for averages)
}

// graphView adapts the Engine's road network to the dijkstra.GraphReader interface.
// It injects live congestion values and optionally penalizes emergency corridor roads.
type graphView struct {
	e              *Engine
	avoidEmergency bool // when true, corridor roads get inflated cost
}

func (g *graphView) Nodes() int { return g.e.cfg.NumIntersections }

// Neighbors returns all outgoing edges from node with live congestion values.
// If avoidEmergency is true, roads on the emergency corridor get artificially
// high congestion (10) and distance (50×) so Dijkstra routes around them.
func (g *graphView) Neighbors(node int) []model.Edge {
	edges := g.e.graph.Adj[node]
	out := make([]model.Edge, 0, len(edges))
	for _, edge := range edges {
		if edge.Road == nil {
			continue
		}
		roadID := edge.Road.ID
		g.e.roadMu.RLock()
		rs := g.e.roadByID[roadID]
		if rs == nil {
			g.e.roadMu.RUnlock()
			continue
		}
		r := rs.road
		cong := rs.congestion
		_, onCorr := g.e.roadOnCorr[roadID]
		g.e.roadMu.RUnlock()

		r.Congestion = cong
		// If routing a normal vehicle and this road is on the emergency corridor,
		// make it prohibitively expensive.
		if g.avoidEmergency && onCorr {
			r.Congestion = 10
			r.Distance = r.Distance * 50
		}
		out = append(out, model.Edge{To: edge.To, Road: &r})
	}
	return out
}

// GetRoad is part of GraphReader. Returns a copy of the road with live congestion.
func (g *graphView) GetRoad(roadID int) (*model.Road, bool) {
	g.e.roadMu.RLock()
	rs := g.e.roadByID[roadID]
	if rs == nil {
		g.e.roadMu.RUnlock()
		return nil, false
	}
	r := rs.road
	cong := rs.congestion
	g.e.roadMu.RUnlock()
	r.Congestion = cong
	return &r, true
}

// --- Vehicle implementation ---

// vehicle holds the mutable state for one vehicle in the simulation.
// Each vehicle runs in its own goroutine via v.run().
type vehicle struct {
	plate string      // unique plate ID
	kind  VehicleType // normal or emergency
	to    int         // destination intersection ID

	mu    sync.Mutex
	state VehicleState // current lifecycle state
	at    int          // current intersection ID (valid when at intersection or waiting)
	path  []int        // sequence of intersection IDs from current position to destination
	idx   int          // current position index into path (path[idx] == v.at when at intersection)

	onRoadID      int // road ID currently traversing (valid when OnRoad)
	remainingRoad int // ticks remaining before arriving at next intersection

	queued      bool   // true if this vehicle is counted in an intersection queue
	queuedGroup string // "NS" or "EW" — which queue group we're in

	processedTick int64 // last tick processed (used by tests via WaitVehicleProcessed)
}

// run is the vehicle goroutine's main loop: wait for a tick, step, record progress.
func (v *vehicle) run(ctx context.Context, e *Engine) {
	var last int64
	for {
		n, ok := e.waitTick(last)
		if !ok {
			return
		}
		last = n
		v.step(e)
		atomic.StoreInt64(&v.processedTick, last)
	}
}

// step runs one tick of the vehicle's state machine. Called from run().
//
// State graph:
//
//	AtIntersection ──(signal green)──→ OnRoad ──(remainingRoad hits 0)──→ AtIntersection
//	       │                                                            │
//	       └──(signal red)──→ WaitingSignal ──(signal green)─────────────┘
//	                                                                     │
//	                              AtIntersection ──(at == to)─────────────→ Arrived
func (v *vehicle) step(e *Engine) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Already arrived — nothing to do.
	if v.state == VehicleStateArrived {
		return
	}

	// On a road: count down remaining ticks. When we hit 0, arrive at next intersection.
	if v.state == VehicleStateOnRoad {
		v.remainingRoad--
		if v.remainingRoad > 0 {
			return
		}
		// Road traversal complete: decrement occupancy for road we're leaving.
		e.roadMu.Lock()
		rs := e.roadByID[v.onRoadID]
		if rs != nil && rs.occupancy > 0 {
			rs.occupancy--
		}
		e.roadMu.Unlock()

		v.onRoadID = 0
		v.at = v.path[v.idx] // idx was incremented when we entered the road
		v.state = VehicleStateAtIntersection
		v.queued = false
		v.queuedGroup = ""
	}

	// Check if we reached the final destination.
	if v.at == v.to {
		v.state = VehicleStateArrived
		return
	}

	if v.idx >= len(v.path)-1 {
		return
	}

	// Get the next hop and check the signal at the current intersection.
	from := v.path[v.idx]
	to := v.path[v.idx+1]
	inter := e.signals[from]
	if inter == nil {
		return
	}

	// Determine movement direction (NS or EW) and check if signal allows it.
	group := movementDirection(from, to)
	if phaseAllows(inter.currentPhase(), group) {
		// Signal is green for our direction — cross the intersection.
		if v.queued {
			inter.dequeue(v.queuedGroup)
			v.queued = false
			v.queuedGroup = ""
		}
		roadID, travelTicks, ok := e.roadForHop(from, to)
		if !ok {
			return
		}
		// Increment occupancy for the road we're entering.
		e.roadMu.Lock()
		rs := e.roadByID[roadID]
		if rs != nil {
			rs.occupancy++
		}
		e.roadMu.Unlock()

		v.onRoadID = roadID
		v.remainingRoad = travelTicks
		v.state = VehicleStateOnRoad
		v.idx++ // advance to next hop; path[v.idx] becomes the destination of this road
		return
	}

	// Signal is red (or yellow) for our direction — wait.
	if !v.queued {
		inter.enqueue(group)
		v.queued = true
		v.queuedGroup = group
	}
	v.state = VehicleStateWaitingSignal
}

func phaseAllows(p Phase, group string) bool {
	switch p {
	case PhaseGreenNS:
		return group == "NS"
	case PhaseGreenEW:
		return group == "EW"
	default:
		return false
	}
}

func (e *Engine) roadForHop(from, to int) (roadID int, travelTicks int, ok bool) {
	for _, edge := range e.graph.Adj[from] {
		if edge.To == to && edge.Road != nil {
			roadID = edge.Road.ID
			e.roadMu.RLock()
			rs := e.roadByID[roadID]
			if rs == nil {
				e.roadMu.RUnlock()
				return 0, 0, false
			}
			r := rs.road
			cong := rs.congestion
			e.roadMu.RUnlock()

			r.Congestion = cong
			w := dijkstra.WeightMinutes(&r)
			travelTicks = int(math.Ceil(w))
			if travelTicks < 1 {
				travelTicks = 1
			}
			return roadID, travelTicks, true
		}
	}
	return 0, 0, false
}

// --- Intersection implementation ---

// intersection is a traffic signal controller at one node.
// It cycles GreenNS → YellowNS → GreenEW → YellowEW → GreenNS.
// Each green duration is adaptive: scales with queue ratio between
// MinGreenTicks and MaxGreenTicks.
type intersection struct {
	id int

	mu            sync.Mutex
	phase         Phase
	remaining     int   // ticks remaining in current phase
	queueNS       int   // vehicles waiting in NS direction
	queueEW       int   // vehicles waiting in EW direction
	preempt       bool  // true when emergency preemption is active
	preemptPhase  Phase // phase to hold during preemption
	preemptRemain int   // ticks of preemption remaining

	cfg Config
}

func newIntersection(id int, cfg Config) *intersection {
	return &intersection{
		id:        id,
		phase:     PhaseGreenNS,
		remaining: cfg.MinGreenTicks,
		cfg:       cfg,
	}
}

// run is the intersection goroutine's main loop: wait for tick, step signal.
func (i *intersection) run(ctx context.Context, e *Engine) {
	var last int64
	for {
		n, ok := e.waitTick(last)
		if !ok {
			return
		}
		last = n
		if ctx.Err() != nil {
			return
		}
		i.step()
	}
}

// step advances the signal phase by one tick.
// Priority: preemption > countdown > phase transition.
func (i *intersection) step() {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Preemption override: hold the preempted phase and count down.
	if i.preempt {
		i.phase = i.preemptPhase
		i.preemptRemain--
		if i.preemptRemain <= 0 {
			i.preempt = false
		}
		return
	}

	// Normal countdown: ticks remaining in current phase.
	i.remaining--
	if i.remaining > 0 {
		return
	}

	// Phase transition: move to next state in the cycle.
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
	default:
		i.phase = PhaseGreenNS
		i.remaining = i.cfg.MinGreenTicks
	}
}

// adaptiveGreenTicks computes green duration based on queue ratios.
// More queued vehicles → longer green, clamped to [MinGreenTicks, MaxGreenTicks].
func (i *intersection) adaptiveGreenTicks(group string) int {
	ns := i.queueNS
	ew := i.queueEW
	total := ns + ew
	if total <= 0 {
		return i.cfg.MinGreenTicks
	}
	var ratio float64
	if group == "NS" {
		ratio = float64(ns) / float64(total)
	} else {
		ratio = float64(ew) / float64(total)
	}
	green := float64(i.cfg.MinGreenTicks) + ratio*float64(i.cfg.MaxGreenTicks-i.cfg.MinGreenTicks)
	out := int(math.Round(green))
	if out < i.cfg.MinGreenTicks {
		out = i.cfg.MinGreenTicks
	}
	if out > i.cfg.MaxGreenTicks {
		out = i.cfg.MaxGreenTicks
	}
	return out
}

// currentPhase returns the current phase (thread-safe).
func (i *intersection) currentPhase() Phase {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.phase
}

// enqueue adds a vehicle to the waiting queue for a direction.
func (i *intersection) enqueue(group string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if group == "NS" {
		i.queueNS++
		return
	}
	i.queueEW++
}

// dequeue removes a vehicle from the waiting queue for a direction.
func (i *intersection) dequeue(group string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if group == "NS" {
		if i.queueNS > 0 {
			i.queueNS--
		}
		return
	}
	if i.queueEW > 0 {
		i.queueEW--
	}
}

// status returns a SignalStatus snapshot for this intersection.
func (i *intersection) status() SignalStatus {
	i.mu.Lock()
	defer i.mu.Unlock()
	return SignalStatus{
		Phase:          i.phase,
		RemainingTicks: i.remaining,
		QueueNS:        i.queueNS,
		QueueEW:        i.queueEW,
		Preempt:        i.preempt,
	}
}

// setPreempt forces the signal to hold a specific green phase for a duration.
// Called by setEmergencyCorridor.
func (i *intersection) setPreempt(phase Phase, ticks int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.preempt = true
	i.preemptPhase = phase
	i.preemptRemain = ticks
}

// Compile-time check that graphView implements GraphReader.
var _ model.GraphReader = (*graphView)(nil)
