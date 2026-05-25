package scenario

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestLoadBasic(t *testing.T) {
	s, err := LoadFile(filepath.Join("..", "..", "scenarios", "01_basic.json"))
	if err != nil {
		t.Fatalf("load 01_basic: %v", err)
	}
	if s.Name != "01_basic" {
		t.Fatalf("expected name 01_basic got %q", s.Name)
	}
	if s.TickMs != 50 {
		t.Fatalf("expected tickMs 50 got %d", s.TickMs)
	}
	if len(s.Intersections) != 4 {
		t.Fatalf("expected 4 intersections got %d", len(s.Intersections))
	}
	if len(s.Roads) != 6 {
		t.Fatalf("expected 6 roads got %d", len(s.Roads))
	}
	if s.Vehicle.Plate != "CAR-01" {
		t.Fatalf("expected CAR-01 got %q", s.Vehicle.Plate)
	}
}

func TestLoadGridShorthand(t *testing.T) {
	s, err := LoadFile(filepath.Join("..", "..", "scenarios", "03_grid.json"))
	if err != nil {
		t.Fatalf("load 03_grid: %v", err)
	}
	if len(s.Intersections) != 20 {
		t.Fatalf("expected 20 intersections got %d", len(s.Intersections))
	}
	if len(s.Roads) != 62 {
		t.Fatalf("expected 62 roads got %d", len(s.Roads))
	}
	// Verify column-major ordering
	x, y, ok := s.Intersections[0].XY()
	if !ok || x != 0 || y != 0 {
		t.Fatalf("expected intersection 0 at (0,0) got (%d,%d)", x, y)
	}
	x, y, ok = s.Intersections[4].XY()
	if !ok || x != 1 || y != 0 {
		t.Fatalf("expected intersection 4 at (1,0) got (%d,%d)", x, y)
	}
	x, y, ok = s.Intersections[19].XY()
	if !ok || x != 4 || y != 3 {
		t.Fatalf("expected intersection 19 at (4,3) got (%d,%d)", x, y)
	}
}

func TestLoadEmergencyCompact(t *testing.T) {
	s, err := LoadFile(filepath.Join("..", "..", "scenarios", "02_emergency_preempt.json"))
	if err != nil {
		t.Fatalf("load 02_emergency_preempt: %v", err)
	}
	if len(s.Events) != 1 {
		t.Fatalf("expected 1 event got %d", len(s.Events))
	}
	if s.Events[0].Type != "emergency" {
		t.Fatalf("expected emergency type got %q", s.Events[0].Type)
	}
	if s.Events[0].Tick != 5 {
		t.Fatalf("expected tick 5 got %d", s.Events[0].Tick)
	}
	// Should have 8 roads (4 compact entries × 2 for bidirectional)
	if len(s.Roads) != 8 {
		t.Fatalf("expected 8 roads (4 compact × 2 bidirectional) got %d", len(s.Roads))
	}
}

func TestBuild(t *testing.T) {
	s, err := LoadFile(filepath.Join("..", "..", "scenarios", "01_basic.json"))
	if err != nil {
		t.Fatal(err)
	}
	built, err := s.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if built.Graph.Nodes != len(s.Intersections) {
		t.Fatalf("expected Nodes=%d got %d", len(s.Intersections), built.Graph.Nodes)
	}
}

func TestParseRoadsCompact(t *testing.T) {
	raw := json.RawMessage(`[[0, 1, 1.0, 30, 2], [1, 2, 1.5, 40, 3]]`)
	roads, err := parseRoads(raw, false)
	if err != nil {
		t.Fatalf("parseRoads: %v", err)
	}
	if len(roads) != 2 {
		t.Fatalf("expected 2 roads got %d", len(roads))
	}
	if roads[0].From != 0 || roads[0].To != 1 || roads[0].DistanceKm != 1.0 || roads[0].SpeedKmph != 30 || roads[0].Congestion != 2 {
		t.Fatalf("road 0 mismatch: %+v", roads[0])
	}
	if roads[1].From != 1 || roads[1].To != 2 || roads[1].DistanceKm != 1.5 || roads[1].SpeedKmph != 40 || roads[1].Congestion != 3 {
		t.Fatalf("road 1 mismatch: %+v", roads[1])
	}
}

func TestParseRoadsCompactBidirectional(t *testing.T) {
	raw := json.RawMessage(`[[0, 1, 1.0, 30, 2]]`)
	roads, err := parseRoads(raw, true)
	if err != nil {
		t.Fatalf("parseRoads: %v", err)
	}
	if len(roads) != 2 {
		t.Fatalf("expected 2 roads (bidirectional) got %d", len(roads))
	}
	if roads[0].From != 0 || roads[0].To != 1 {
		t.Fatalf("forward: expected 0→1 got %d→%d", roads[0].From, roads[0].To)
	}
	if roads[1].From != 1 || roads[1].To != 0 {
		t.Fatalf("reverse: expected 1→0 got %d→%d", roads[1].From, roads[1].To)
	}
}

func TestParseRoadsVerboseBidirectional(t *testing.T) {
	raw := json.RawMessage(`[{"from":0,"to":1,"distanceKm":1,"speedKmph":30,"congestion":2}]`)
	roads, err := parseRoads(raw, true)
	if err != nil {
		t.Fatalf("parseRoads: %v", err)
	}
	if len(roads) != 2 {
		t.Fatalf("expected 2 roads (bidirectional) got %d", len(roads))
	}
	if roads[0].From != 0 || roads[0].To != 1 {
		t.Fatalf("forward: expected 0→1")
	}
	if roads[1].From != 1 || roads[1].To != 0 {
		t.Fatalf("reverse: expected 1→0")
	}
}

func TestGenerateGrid(t *testing.T) {
	gc := &GridConfig{Cols: 3, Rows: 2, CellDist: 1.0, Speed: 30, Congestion: 2}
	inters, roads := generateGrid(gc)
	if len(inters) != 6 {
		t.Fatalf("expected 6 intersections got %d", len(inters))
	}
	// Column-major: id = col*rows + row
	// col 0: id 0,1; col 1: id 2,3; col 2: id 4,5
	type check struct{ id, x, y int }
	for _, c := range []check{{0, 0, 0}, {1, 0, 1}, {2, 1, 0}, {3, 1, 1}, {4, 2, 0}, {5, 2, 1}} {
		if inters[c.id].ID != c.id || *inters[c.id].X != c.x || *inters[c.id].Y != c.y {
			t.Fatalf("intersection %d: expected (%d,%d) got (%d,%d)", c.id, c.x, c.y, *inters[c.id].X, *inters[c.id].Y)
		}
	}
	// Roads: 3 cols × 1 vertical each × 2 directions = 6 + 2 cols × 2 rows horizontal × 2 directions = 8 = 14 total
	expectedRoads := (gc.Cols*(gc.Rows-1) + (gc.Cols-1)*gc.Rows) * 2
	if len(roads) != expectedRoads {
		t.Fatalf("expected %d roads got %d", expectedRoads, len(roads))
	}
}

func TestValidationFailures(t *testing.T) {
	tests := []struct {
		name string
		s    Scenario
	}{
		{"missing tickMs", Scenario{}},
		{"empty intersections", Scenario{TickMs: 100}},
		{"empty roads", Scenario{TickMs: 100, Intersections: []Intersection{{ID: 0, X: intPtr(0), Y: intPtr(0)}}}},
		{"no vehicle", Scenario{
			TickMs:        100,
			Intersections: []Intersection{{ID: 0, X: intPtr(0), Y: intPtr(0)}},
			Roads:         []Road{{ID: 1, From: 0, To: 0, DistanceKm: 1, SpeedKmph: 30, Congestion: 2}},
		}},
		{"bad vehicle type", Scenario{
			TickMs:        100,
			Intersections: []Intersection{{ID: 0, X: intPtr(0), Y: intPtr(0)}},
			Roads:         []Road{{ID: 1, From: 0, To: 0, DistanceKm: 1, SpeedKmph: 30, Congestion: 2}},
			Vehicle:       Vehicle{Plate: "X", Type: "invalid", From: 0, To: 0},
		}},
		{"missing intersection", Scenario{
			TickMs:        100,
			Intersections: []Intersection{{ID: 0, X: intPtr(0), Y: intPtr(0)}},
			Roads:         []Road{{ID: 1, From: 0, To: 1, DistanceKm: 1, SpeedKmph: 30, Congestion: 2}},
			Vehicle:       Vehicle{Plate: "X", Type: "normal", From: 0, To: 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Validate(); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func intPtr(v int) *int { return &v }

func TestBuildNonContiguous(t *testing.T) {
	s := Scenario{
		TickMs: 100,
		Intersections: []Intersection{
			{ID: 0, X: intPtr(0), Y: intPtr(0)},
			{ID: 2, X: intPtr(1), Y: intPtr(0)},
		},
		Roads: []Road{
			{ID: 0, From: 0, To: 2, DistanceKm: 1, SpeedKmph: 30, Congestion: 2},
			{ID: 1, From: 2, To: 0, DistanceKm: 1, SpeedKmph: 30, Congestion: 2},
		},
		Vehicle: Vehicle{Plate: "X", Type: "normal", From: 0, To: 2},
	}
	if _, err := s.Build(); err == nil {
		t.Fatal("expected error for non-contiguous ids")
	}
}
