package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/engine"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/scenario"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/sim"
)

func main() {
	scenarioPath := flag.String("scenario", "", "path to scenario JSON (required)")
	maxTicks := flag.Int64("max-ticks", 5000, "max ticks to run before failing")
	flag.Parse()

	if *scenarioPath == "" {
		fmt.Fprintln(os.Stderr, "-scenario is required")
		flag.Usage()
		os.Exit(2)
	}

	// Load and build the scenario.
	s, err := scenario.LoadFile(*scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load scenario: %v\n", err)
		os.Exit(2)
	}
	built, err := s.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build scenario: %v\n", err)
		os.Exit(2)
	}

	cfg := engine.DefaultConfig()
	cfg.NumIntersections = built.Graph.Nodes
	cfg.TickDuration = s.TickDuration()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create engine with a real clock (wall-clock ticks).
	clock := sim.NewRealClock(ctx, cfg.TickDuration)
	e, err := engine.New(cfg, built.Graph, clock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine: %v\n", err)
		os.Exit(1)
	}
	e.Start(ctx)
	defer e.Stop()

	// Register the scenario's vehicle.
	vehType := engine.VehicleTypeNormal
	if s.Vehicle.Type == "emergency" {
		vehType = engine.VehicleTypeEmergency
	}
	routeResp, err := e.RegisterVehicle(ctx, engine.RegisterVehicleRequest{
		Plate: s.Vehicle.Plate, From: s.Vehicle.From, To: s.Vehicle.To, Type: vehType,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "register vehicle: %v\n", err)
		os.Exit(1)
	}

	// Print static info (overview + ASCII map).
	printOverview(s, routeResp)
	printMap(built, s, routeResp.Path)

	// Track last-known state to detect changes (state-change-only output).
	var lastState struct {
		state       engine.VehicleState
		at, next    int
		signalPhase engine.Phase
		preempt     bool
	}
	var evIdx int
	var tick int64

	// Simulation loop: wait for each tick, check events, print state changes.
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var ok bool
		tick, ok = e.WaitNextTick(tick)
		if !ok {
			return
		}
		if tick > *maxTicks {
			fmt.Fprintf(os.Stderr, "ERROR: vehicle did not arrive within %d ticks (ETA=%.1fmin)\n", *maxTicks, routeResp.ETAMinutes)
			os.Exit(1)
		}

		// Fire any scheduled events at this tick.
		for evIdx < len(s.Events) && s.Events[evIdx].Tick == tick {
			ev := s.Events[evIdx]
			if ev.Type == "emergency" {
				fmt.Printf("--- tick %d: EMERGENCY %s (%d\u2192%d)\n", tick, ev.Plate, ev.From, ev.To)
				resp, err := e.DispatchEmergency(ctx, engine.DispatchEmergencyRequest{
					Plate: ev.Plate, From: ev.From, To: ev.To,
				})
				if err == nil {
					fmt.Printf("    route %v, preempting %d intersections\n", resp.Path, len(resp.Path)-1)
				}
			}
			evIdx++
		}

		// Read current vehicle state.
		st, err := e.VehicleStatus(ctx, s.Vehicle.Plate)
		if err != nil {
			continue
		}

		// Detect state change.
		changed := st.State != lastState.state ||
			st.At != lastState.at ||
			st.Next != lastState.next

		// Suppress repetitive "on_road" and "arrived" lines (only print once).
		if !changed && (st.State == engine.VehicleStateOnRoad || st.State == engine.VehicleStateArrived) {
			continue
		}

		// At intersections, also check signal phase/preempt changes.
		if st.State == engine.VehicleStateAtIntersection || st.State == engine.VehicleStateWaitingSignal {
			sig, _ := e.SignalStatus(ctx, st.At)
			if !changed && sig.Phase == lastState.signalPhase && sig.Preempt == lastState.preempt {
				continue
			}
			lastState.signalPhase = sig.Phase
			lastState.preempt = sig.Preempt
		}

		lastState.state = st.State
		lastState.at = st.At
		lastState.next = st.Next

		// Print one-line summary for this tick.
		switch st.State {
		case engine.VehicleStateOnRoad:
			fmt.Printf("t=%-5d road %d\u2192%d rem=%-3d\n", tick, st.At, st.Next, st.RemainingRoadTicks)
		case engine.VehicleStateAtIntersection:
			fmt.Printf("t=%-5d at %d\n", tick, st.At)
		case engine.VehicleStateWaitingSignal:
			sig, _ := e.SignalStatus(ctx, st.At)
			pre := ""
			if sig.Preempt {
				pre = " [preempt]"
			}
			fmt.Printf("t=%-5d wait %d \u2190%s rem=%-3d%s\n", tick, st.At, sig.Phase, sig.RemainingTicks, pre)
		case engine.VehicleStateArrived:
			fmt.Printf("t=%-5d arrived actual=%d ticks (ETA %.1fmin)\n", tick, tick, routeResp.ETAMinutes)
			return
		}
	}
}

// printOverview prints the scenario name, vehicle info, route, and events.
func printOverview(s *scenario.Scenario, routeResp engine.RouteResponse) {
	name := s.Name
	if name == "" {
		name = "(unnamed)"
	}
	fmt.Printf("Scenario: %s\n", name)
	fmt.Printf("Tick: %dms\n", s.TickMs)
	fmt.Printf("Vehicle: %s (%s) from=%d to=%d\n", s.Vehicle.Plate, s.Vehicle.Type, s.Vehicle.From, s.Vehicle.To)
	fmt.Printf("Route: %v\n", routeResp.Path)
	if len(s.Events) > 0 {
		fmt.Printf("Events:\n")
		for _, ev := range s.Events {
			fmt.Printf("  t=%d  %s %s %d\u2192%d\n", ev.Tick, ev.Type, ev.Plate, ev.From, ev.To)
		}
	}
	fmt.Println()
}

type roadDraw struct {
	from, to int
	roadID   int
}

// printMap renders an ASCII map of the road network with intersections and roads.
// Horizontal roads use ─, vertical roads use │.
func printMap(b *scenario.Built, s *scenario.Scenario, path []int) {
	// Compute bounding box.
	minX, minY := int(^uint(0)>>1), int(^uint(0)>>1)
	maxX, maxY := -minX, -minY
	for id := range b.Intersections {
		x, y, ok := b.XY(id)
		if !ok {
			continue
		}
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}
	if (maxX-minX) > 200 || (maxY-minY) > 200 {
		fmt.Printf("Map: (%d\u00d7%d) too large to render\n\n", maxX-minX+1, maxY-minY+1)
		return
	}

	// Determine cell width from the widest intersection label.
	cellW := 2
	for _, id := range path {
		if d := len(fmt.Sprintf("%d", id)); d > cellW {
			cellW = d
		}
	}
	cellW += 1

	// Classify roads as horizontal or vertical for drawing.
	drawH := make([]roadDraw, 0)
	drawV := make([]roadDraw, 0)
	skipped := 0
	for _, r := range s.Roads {
		x1, y1, ok1 := b.XY(r.From)
		x2, y2, ok2 := b.XY(r.To)
		if !ok1 || !ok2 {
			skipped++
			continue
		}
		if y1 == y2 && x1 != x2 {
			drawH = append(drawH, roadDraw{from: r.From, to: r.To, roadID: r.ID})
		} else if x1 == x2 && y1 != y2 {
			drawV = append(drawV, roadDraw{from: r.From, to: r.To, roadID: r.ID})
		} else {
			skipped++
		}
	}

	cellH := 1
	if len(drawV) > 0 {
		cellH = 2 // extra row for vertical connectors
	}
	w := (maxX - minX + 1) * cellW
	h := (maxY-minY)*cellH + 1
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	canvas := make([][]rune, h)
	for i := 0; i < h; i++ {
		row := make([]rune, w)
		for j := range row {
			row[j] = ' '
		}
		canvas[i] = row
	}

	// Draw intersection labels.
	idLabel := make(map[int]string)
	ids := make([]int, 0, len(b.Intersections))
	for id := range b.Intersections {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		x, y, _ := b.XY(id)
		cx := (x - minX) * cellW
		cy := (y - minY) * cellH
		label := fmt.Sprintf("%d", id)
		idLabel[id] = label
		for i := 0; i < len(label) && cx+i < w; i++ {
			canvas[cy][cx+i] = rune(label[i])
		}
	}

	// Draw horizontal roads (─).
	for _, rd := range drawH {
		x1, y1, _ := b.XY(rd.from)
		x2, _, _ := b.XY(rd.to)
		y := (y1 - minY) * cellH
		cx1 := (x1 - minX) * cellW
		cx2 := (x2 - minX) * cellW
		if cx1 > cx2 {
			cx1, cx2 = cx2, cx1
		}
		start := cx1 + len(idLabel[rd.from])
		if start < cx2 {
			for x := start; x < cx2 && x < w; x++ {
				if canvas[y][x] == ' ' {
					canvas[y][x] = '\u2500'
				}
			}
		}
	}

	// Draw vertical roads (│).
	for _, rd := range drawV {
		x1, y1, _ := b.XY(rd.from)
		_, y2, _ := b.XY(rd.to)
		x := (x1 - minX) * cellW
		cy1 := (y1 - minY) * cellH
		cy2 := (y2 - minY) * cellH
		if cy1 > cy2 {
			cy1, cy2 = cy2, cy1
		}
		for y := cy1 + 1; y < cy2 && y < h; y++ {
			if canvas[y][x] == ' ' {
				canvas[y][x] = '\u2502'
			}
		}
	}

	fmt.Println("Map:")
	for i := 0; i < h; i++ {
		fmt.Println(string(canvas[i]))
	}
	if len(path) > 0 {
		hops := make([]string, 0, len(path)-1)
		for i := 0; i < len(path)-1; i++ {
			hops = append(hops, fmt.Sprintf("%d\u2192%d", path[i], path[i+1]))
		}
		fmt.Printf("Path: %s\n", strings.Join(hops, " "))
	}
	if skipped > 0 {
		fmt.Printf("Skipped %d non-axis-aligned road(s)\n", skipped)
	}
	if len(path) > 0 {
		fmt.Printf("S=%d  D=%d\n\n", path[0], path[len(path)-1])
	}
}
