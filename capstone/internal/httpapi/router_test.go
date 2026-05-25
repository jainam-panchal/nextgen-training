package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/engine"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/sim"
)

func TestHealthz(t *testing.T) {
	cfg := engine.DefaultConfig()
	// Keep greens >1 tick so vehicle doesn't miss green due to goroutine scheduling.
	cfg.MinGreenTicks = 3
	cfg.MaxGreenTicks = 3
	cfg.YellowTicks = 1
	cfg.HistoryWindow = 10

	clock := sim.NewManualClock()
	e, err := engine.New(cfg, model.GenerateDemoGraph(), clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	s := New(e)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestRegisterVehicleAndRoute(t *testing.T) {
	cfg := engine.DefaultConfig()
	// Keep greens >1 tick so vehicle doesn't miss green due to goroutine scheduling.
	cfg.MinGreenTicks = 3
	cfg.MaxGreenTicks = 3
	cfg.YellowTicks = 1
	cfg.HistoryWindow = 10

	clock := sim.NewManualClock()
	e, err := engine.New(cfg, model.GenerateDemoGraph(), clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	s := New(e)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body, _ := json.Marshal(engine.RegisterVehicleRequest{Plate: "KA01", From: 0, To: 3, Type: engine.VehicleTypeNormal})
	resp, err := http.Post(ts.URL+"/vehicles", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	// Advance simulation ticks so the vehicle can move.
	// Each hop is ~2 minutes (weight units), so give it enough ticks.
	for i := 0; i < 50; i++ {
		clock.Step()
		e.WaitVehicleProcessed("KA01", int64(i+1))
	}
	deadline := time.Now().Add(250 * time.Millisecond)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for vehicle to arrive")
		}
		r, err := http.Get(ts.URL + "/stats")
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		var st engine.Stats
		if err := json.Unmarshal(b, &st); err != nil {
			t.Fatalf("decode stats: %v body=%s", err, string(b))
		}
		if st.VehiclesTotal == 1 && st.VehiclesArrived >= 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	resp, err = http.Get(ts.URL + "/route?from=0&to=3")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}

func TestEmergencyDispatchPreemptsSignals(t *testing.T) {
	cfg := engine.DefaultConfig()
	cfg.MinGreenTicks = 3
	cfg.MaxGreenTicks = 3
	cfg.YellowTicks = 1
	cfg.HistoryWindow = 10
	cfg.PreemptionTicks = 20

	clock := sim.NewManualClock()
	e, err := engine.New(cfg, model.GenerateDemoGraph(), clock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.Start(ctx)
	defer e.Stop()

	s := New(e)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body, _ := json.Marshal(engine.DispatchEmergencyRequest{Plate: "AMB-9", From: 0, To: 19})
	resp, err := http.Post(ts.URL+"/emergency", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}

	clock.Step()
	// Intersection 0 should be preempted to a green phase.
	r, err := http.Get(ts.URL + "/signals/0")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	if r.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", r.StatusCode, string(b))
	}
	var st engine.SignalStatus
	if err := json.Unmarshal(b, &st); err != nil {
		t.Fatalf("decode signal: %v body=%s", err, string(b))
	}
	if !st.Preempt {
		t.Fatalf("expected preempt true")
	}
	if st.Phase != engine.PhaseGreenNS && st.Phase != engine.PhaseGreenEW {
		t.Fatalf("expected green phase, got %s", st.Phase)
	}
}
