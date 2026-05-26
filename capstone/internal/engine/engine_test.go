package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/sim"
)

func TestSignalPhaseOrder(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MinGreenTicks = 2
	cfg.MaxGreenTicks = 2
	cfg.YellowTicks = 1
	cfg.HistoryWindow = 10

	clock := sim.NewManualClock()
	e, err := New(cfg, model.GenerateDemoGraph(), clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	// Initial phase is Green-NS.
	st, err := e.SignalStatus(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if st.Phase != PhaseGreenNS {
		t.Fatalf("expected %s got %s", PhaseGreenNS, st.Phase)
	}

	// Step enough ticks to observe full cycle.
	for i := 0; i < 20; i++ {
		clock.Step()
	}
	// Assert phase is one of defined phases.
	st, _ = e.SignalStatus(context.Background(), 0)
	switch st.Phase {
	case PhaseGreenNS, PhaseYellowNS, PhaseGreenEW, PhaseYellowEW:
		// ok
	default:
		t.Fatalf("unexpected phase %q", st.Phase)
	}
}

func TestEmergencyPreemptionForcesGreen(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MinGreenTicks = 5
	cfg.MaxGreenTicks = 5
	cfg.YellowTicks = 2
	cfg.PreemptionTicks = 5
	cfg.HistoryWindow = 10

	clock := sim.NewManualClock()
	e, err := New(cfg, model.GenerateDemoGraph(), clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	// Dispatch emergency along increasing IDs => should preempt Green-NS on intersections on the path.
	_, err = e.DispatchEmergency(context.Background(), DispatchEmergencyRequest{Plate: "AMB-1", From: 0, To: 3})
	if err != nil {
		t.Fatal(err)
	}

	clock.Step()
	st, err := e.SignalStatus(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Preempt {
		t.Fatalf("expected preempt true")
	}
	if st.Phase != PhaseGreenNS {
		t.Fatalf("expected %s got %s", PhaseGreenNS, st.Phase)
	}
}

func TestLoad100VehiclesRaceSafe(t *testing.T) {
	// This is primarily a race/dedlock smoke test. Run with -race.
	cfg := DefaultConfig()
	cfg.MinGreenTicks = 1
	cfg.MaxGreenTicks = 1
	cfg.YellowTicks = 1
	cfg.HistoryWindow = 10

	clock := sim.NewManualClock()
	e, err := New(cfg, model.GenerateDemoGraph(), clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			plate := "V" + itoa(i)
			_, _ = e.RegisterVehicle(context.Background(), RegisterVehicleRequest{Plate: plate, From: 0, To: 19, Type: VehicleTypeNormal})
		}(i)
	}
	wg.Wait()

	// Run some ticks.
	for i := 0; i < 200; i++ {
		clock.Step()
	}

	// Give goroutines a moment to consume ticks.
	time.Sleep(10 * time.Millisecond)

	stats := e.Stats(context.Background())
	if stats.VehiclesTotal != 100 {
		t.Fatalf("expected 100 vehicles, got %d", stats.VehiclesTotal)
	}
}

func TestVehicleStatus(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MinGreenTicks = 3
	cfg.MaxGreenTicks = 3
	cfg.YellowTicks = 1
	cfg.HistoryWindow = 10

	clock := sim.NewManualClock()
	e, err := New(cfg, model.GenerateDemoGraph(), clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	_, err = e.RegisterVehicle(context.Background(), RegisterVehicleRequest{Plate: "T1", From: 0, To: 3, Type: VehicleTypeNormal})
	if err != nil {
		t.Fatal(err)
	}

	// Step one tick, let vehicle process.
	clock.Step()

	st, err := e.VehicleStatus(context.Background(), "T1")
	if err != nil {
		t.Fatalf("VehicleStatus: %v", err)
	}
	if st.Plate != "T1" {
		t.Fatalf("expected T1 got %s", st.Plate)
	}
	if st.Type != VehicleTypeNormal {
		t.Fatalf("expected normal got %s", st.Type)
	}
	if st.At != 0 {
		t.Fatalf("expected at=0 got %d", st.At)
	}
	if st.To != 3 {
		t.Fatalf("expected to=3 got %d", st.To)
	}
}

func TestScenarioArrival(t *testing.T) {
	// Chain with varied distance, speed, congestion per segment
	g := model.NewGraph(4)
	g.AddEdge(0, 1, &model.Road{ID: 0, From: 0, To: 1, Distance: 1.0, SpeedLimit: 30.0, Congestion: 3})
	g.AddEdge(1, 2, &model.Road{ID: 1, From: 1, To: 2, Distance: 1.0, SpeedLimit: 60.0, Congestion: 1})
	g.AddEdge(2, 3, &model.Road{ID: 2, From: 2, To: 3, Distance: 2.0, SpeedLimit: 45.0, Congestion: 2})
	g.AddEdge(1, 0, &model.Road{ID: 100, From: 1, To: 0, Distance: 1.0, SpeedLimit: 30.0, Congestion: 3})
	g.AddEdge(2, 1, &model.Road{ID: 101, From: 2, To: 1, Distance: 1.0, SpeedLimit: 60.0, Congestion: 1})
	g.AddEdge(3, 2, &model.Road{ID: 102, From: 3, To: 2, Distance: 2.0, SpeedLimit: 45.0, Congestion: 2})

	cfg := DefaultConfig()
	cfg.NumIntersections = 4
	cfg.MinGreenTicks = 1
	cfg.MaxGreenTicks = 1
	cfg.YellowTicks = 1
	cfg.HistoryWindow = 10

	clock := sim.NewManualClock()
	e, err := New(cfg, g, clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	_, err = e.RegisterVehicle(context.Background(), RegisterVehicleRequest{Plate: "C1", From: 0, To: 3, Type: VehicleTypeNormal})
	if err != nil {
		t.Fatal(err)
	}

	// Step ticks until arrived.
	limit := int64(500)
	for n := int64(1); n <= limit; n++ {
		clock.Step()
		e.WaitVehicleProcessed("C1", n)
		st, err := e.VehicleStatus(ctx, "C1")
		if err != nil {
			t.Fatalf("VehicleStatus at tick %d: %v", n, err)
		}
		if st.State == VehicleStateArrived {
			return // success
		}
	}
	t.Fatalf("vehicle did not arrive within %d ticks", limit)
}

func TestScenarioEmergencyPreempt(t *testing.T) {
	g := model.NewGraph(4)
	g.AddEdge(0, 1, &model.Road{ID: 0, From: 0, To: 1, Distance: 5.0, SpeedLimit: 30.0, Congestion: 3})
	g.AddEdge(1, 2, &model.Road{ID: 1, From: 1, To: 2, Distance: 3.0, SpeedLimit: 50.0, Congestion: 1})
	g.AddEdge(2, 3, &model.Road{ID: 2, From: 2, To: 3, Distance: 1.0, SpeedLimit: 40.0, Congestion: 2})
	g.AddEdge(1, 0, &model.Road{ID: 100, From: 1, To: 0, Distance: 5.0, SpeedLimit: 30.0, Congestion: 3})
	g.AddEdge(2, 1, &model.Road{ID: 101, From: 2, To: 1, Distance: 3.0, SpeedLimit: 50.0, Congestion: 1})
	g.AddEdge(3, 2, &model.Road{ID: 102, From: 3, To: 2, Distance: 1.0, SpeedLimit: 40.0, Congestion: 2})

	cfg := DefaultConfig()
	cfg.NumIntersections = 4
	cfg.MinGreenTicks = 10
	cfg.MaxGreenTicks = 10
	cfg.YellowTicks = 3
	cfg.PreemptionTicks = 20
	cfg.HistoryWindow = 10

	clock := sim.NewManualClock()
	e, err := New(cfg, g, clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	// Register normal vehicle
	_, err = e.RegisterVehicle(context.Background(), RegisterVehicleRequest{Plate: "C1", From: 0, To: 3, Type: VehicleTypeNormal})
	if err != nil {
		t.Fatal(err)
	}

	// Step a few ticks
	for i := 0; i < 5; i++ {
		clock.Step()
	}

	// Dispatch emergency
	_, err = e.DispatchEmergency(ctx, DispatchEmergencyRequest{Plate: "AMB-1", From: 0, To: 3})
	if err != nil {
		t.Fatal(err)
	}

	clock.Step()

	st, err := e.SignalStatus(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Preempt {
		t.Fatalf("expected intersection 0 preempted after emergency, got preempt=%v phase=%s", st.Preempt, st.Phase)
	}
}

func itoa(i int) string {
	// tiny helper to avoid strconv import in tests.
	if i == 0 {
		return "0"
	}
	n := i
	if n < 0 {
		n = -n
	}
	var buf [32]byte
	j := len(buf)
	for n > 0 {
		j--
		buf[j] = byte('0' + n%10)
		n /= 10
	}
	if i < 0 {
		j--
		buf[j] = '-'
	}
	return string(buf[j:])
}
