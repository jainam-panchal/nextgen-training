package scenario

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
)

// GridConfig describes a rectangular grid of intersections.
// IDs are assigned column-major: id = col*rows + row (so adjacent IDs are vertical neighbours).
type GridConfig struct {
	Cols       int     `json:"cols"`       // number of columns (x)
	Rows       int     `json:"rows"`       // number of rows (y)
	CellDist   float64 `json:"cellDist"`   // distance between adjacent intersections (km)
	Speed      float64 `json:"speed"`      // speed limit for all roads (km/h)
	Congestion int     `json:"congestion"` // congestion level for all roads (1..10)
}

// Scenario is a complete simulation scenario loaded from JSON.
type Scenario struct {
	Name          string         `json:"name"`          // human-readable label
	TickMs        int            `json:"tickMs"`        // milliseconds per simulation tick
	Bidirectional bool           `json:"bidirectional"` // if true, auto-add reverse roads
	Intersections []Intersection `json:"intersections"` // list of intersections
	Roads         []Road         `json:"-"`             // parsed roads (populated from json.RawMessage or grid)
	Grid          *GridConfig    `json:"grid"`          // grid shorthand (alternative to listing intersections + roads)
	Vehicle       Vehicle        `json:"vehicle"`       // the vehicle to simulate
	Events        []Event        `json:"events"`        // scheduled events (e.g. emergency dispatch)
}

// Intersection is a node in the road graph with mandatory (x,y) coordinates.
type Intersection struct {
	ID int  `json:"id"`
	X  *int `json:"x"` // grid column (must be non-nil when using verbose format)
	Y  *int `json:"y"` // grid row (must be non-nil)
}

// Road is a directed edge between two intersections.
type Road struct {
	ID         int     `json:"id"`
	From       int     `json:"from"`
	To         int     `json:"to"`
	DistanceKm float64 `json:"distanceKm"`
	SpeedKmph  float64 `json:"speedKmph"`
	Congestion int     `json:"congestion"` // 1..10
}

// Vehicle describes the single vehicle in this scenario.
type Vehicle struct {
	Plate string `json:"plate"`
	Type  string `json:"type"` // "normal" or "emergency"
	From  int    `json:"from"`
	To    int    `json:"to"`
}

// Event is a scheduled action that fires at a specific tick.
type Event struct {
	Tick  int64  `json:"tick"`
	Type  string `json:"type"` // "emergency"
	Plate string `json:"plate"`
	From  int    `json:"from"`
	To    int    `json:"to"`
}

// rawScenario is the intermediate JSON-unmarshal struct that keeps Roads as RawMessage
// so we can auto-detect the format (verbose objects, compact arrays, or grid).
type rawScenario struct {
	Name          string          `json:"name"`
	TickMs        int             `json:"tickMs"`
	Bidirectional bool            `json:"bidirectional"`
	Intersections []Intersection  `json:"intersections"`
	Roads         json.RawMessage `json:"roads"`
	Grid          *GridConfig     `json:"grid"`
	Vehicle       Vehicle         `json:"vehicle"`
	Events        []Event         `json:"events"`
}

// XY returns the coordinates of this intersection if both are set.
func (in Intersection) XY() (int, int, bool) {
	if in.X == nil || in.Y == nil {
		return 0, 0, false
	}
	return *in.X, *in.Y, true
}

// LoadFile reads a JSON scenario file, parses roads (auto-detecting format),
// validates, and returns a Scenario. Supports three road formats:
//   - Verbose: objects with id, from, to, distanceKm, speedKmph, congestion
//   - Compact: [[from, to, dist, speed, cong], ...]
//   - Grid shorthand (via GridConfig)
func LoadFile(path string) (*Scenario, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw rawScenario
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	s := &Scenario{
		Name:          raw.Name,
		TickMs:        raw.TickMs,
		Bidirectional: raw.Bidirectional,
		Vehicle:       raw.Vehicle,
		Events:        raw.Events,
	}

	// Grid shorthand generates intersections + roads from dimensions.
	if raw.Grid != nil {
		s.Intersections, s.Roads = generateGrid(raw.Grid)
	} else {
		s.Intersections = raw.Intersections
		if len(raw.Roads) > 0 {
			roads, err := parseRoads(raw.Roads, raw.Bidirectional)
			if err != nil {
				return nil, fmt.Errorf("roads: %w", err)
			}
			s.Roads = roads
		}
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}
	sort.Slice(s.Events, func(i, j int) bool { return s.Events[i].Tick < s.Events[j].Tick })
	return s, nil
}

// generateGrid creates intersections and roads for a rectangular grid.
// IDs are column-major: intersection id = col*rows + row.
// Roads are generated for vertical (NS) and horizontal (EW) adjacencies, both directions.
func generateGrid(g *GridConfig) ([]Intersection, []Road) {
	var intersections []Intersection
	for col := 0; col < g.Cols; col++ {
		for row := 0; row < g.Rows; row++ {
			id := col*g.Rows + row
			x, y := col, row
			intersections = append(intersections, Intersection{ID: id, X: &x, Y: &y})
		}
	}

	var roads []Road
	nextID := 1
	// Vertical roads (NS): same column, adjacent rows
	for col := 0; col < g.Cols; col++ {
		for row := 0; row < g.Rows-1; row++ {
			from := col*g.Rows + row
			to := from + 1
			roads = append(roads, Road{ID: nextID, From: from, To: to, DistanceKm: g.CellDist, SpeedKmph: g.Speed, Congestion: g.Congestion})
			nextID++
			roads = append(roads, Road{ID: nextID, From: to, To: from, DistanceKm: g.CellDist, SpeedKmph: g.Speed, Congestion: g.Congestion})
			nextID++
		}
	}
	// Horizontal roads (EW): same row, adjacent columns
	for col := 0; col < g.Cols-1; col++ {
		for row := 0; row < g.Rows; row++ {
			from := col*g.Rows + row
			to := from + g.Rows
			roads = append(roads, Road{ID: nextID, From: from, To: to, DistanceKm: g.CellDist, SpeedKmph: g.Speed, Congestion: g.Congestion})
			nextID++
			roads = append(roads, Road{ID: nextID, From: to, To: from, DistanceKm: g.CellDist, SpeedKmph: g.Speed, Congestion: g.Congestion})
			nextID++
		}
	}
	return intersections, roads
}

// parseRoads attempts to parse json.RawMessage first as verbose []Road objects,
// then as compact [[from,to,dist,speed,cong], ...] arrays.
func parseRoads(raw json.RawMessage, bidirectional bool) ([]Road, error) {
	// Try verbose format: []Road (objects)
	var roads []Road
	if err := json.Unmarshal(raw, &roads); err == nil {
		for i := range roads {
			if roads[i].ID == 0 {
				roads[i].ID = i + 1
			}
		}
		if bidirectional {
			roads = appendBidirectional(roads)
		}
		return roads, nil
	}

	// Try compact format: [[from, to, dist, speed, cong], ...]
	roads = nil // reset: verbose attempt may have partially populated
	var compact [][5]float64
	if err := json.Unmarshal(raw, &compact); err != nil {
		return nil, fmt.Errorf("expected road objects or [from,to,dist,speed,cong] arrays")
	}

	nextID := 1
	for _, r := range compact {
		roads = append(roads, Road{
			ID:         nextID,
			From:       int(r[0]),
			To:         int(r[1]),
			DistanceKm: r[2],
			SpeedKmph:  r[3],
			Congestion: int(r[4]),
		})
		nextID++
		if bidirectional {
			roads = append(roads, Road{
				ID:         nextID,
				From:       int(r[1]),
				To:         int(r[0]),
				DistanceKm: r[2],
				SpeedKmph:  r[3],
				Congestion: int(r[4]),
			})
			nextID++
		}
	}
	return roads, nil
}

// appendBidirectional creates reverse copies of all roads for bidirectional flag.
func appendBidirectional(roads []Road) []Road {
	used := make(map[int]bool)
	for _, r := range roads {
		used[r.ID] = true
	}
	nextID := len(roads) + 1
	for _, r := range roads {
		for used[nextID] {
			nextID++
		}
		roads = append(roads, Road{
			ID:         nextID,
			From:       r.To,
			To:         r.From,
			DistanceKm: r.DistanceKm,
			SpeedKmph:  r.SpeedKmph,
			Congestion: r.Congestion,
		})
		nextID++
	}
	return roads
}

// TickDuration converts the scenario tickMs to a time.Duration.
func (s *Scenario) TickDuration() time.Duration {
	return time.Duration(s.TickMs) * time.Millisecond
}

// Validate checks all scenario fields for consistency and required values.
func (s *Scenario) Validate() error {
	if s.TickMs <= 0 {
		return errors.New("tickMs is required and must be > 0")
	}
	if len(s.Intersections) == 0 {
		return errors.New("intersections is required")
	}
	if len(s.Roads) == 0 {
		return errors.New("roads is required")
	}
	if s.Vehicle.Plate == "" {
		return errors.New("vehicle.plate is required")
	}
	if s.Vehicle.Type == "" {
		return errors.New("vehicle.type is required")
	}
	if s.Vehicle.Type != "normal" && s.Vehicle.Type != "emergency" {
		return fmt.Errorf("vehicle.type must be normal or emergency, got %q", s.Vehicle.Type)
	}

	byID := make(map[int]Intersection, len(s.Intersections))
	for _, in := range s.Intersections {
		if _, ok := byID[in.ID]; ok {
			return fmt.Errorf("duplicate intersection id %d", in.ID)
		}
		if in.X == nil || in.Y == nil {
			return fmt.Errorf("intersection %d x,y are required", in.ID)
		}
		byID[in.ID] = in
	}
	if _, ok := byID[s.Vehicle.From]; !ok {
		return fmt.Errorf("vehicle.from intersection %d not found", s.Vehicle.From)
	}
	if _, ok := byID[s.Vehicle.To]; !ok {
		return fmt.Errorf("vehicle.to intersection %d not found", s.Vehicle.To)
	}

	roadIDs := make(map[int]struct{}, len(s.Roads))
	for _, r := range s.Roads {
		if _, ok := roadIDs[r.ID]; ok {
			return fmt.Errorf("duplicate road id %d", r.ID)
		}
		roadIDs[r.ID] = struct{}{}
		if _, ok := byID[r.From]; !ok {
			return fmt.Errorf("road %d from intersection %d not found", r.ID, r.From)
		}
		if _, ok := byID[r.To]; !ok {
			return fmt.Errorf("road %d to intersection %d not found", r.ID, r.To)
		}
		if r.DistanceKm <= 0 {
			return fmt.Errorf("road %d distanceKm must be > 0", r.ID)
		}
		if r.SpeedKmph <= 0 {
			return fmt.Errorf("road %d speedKmph must be > 0", r.ID)
		}
		if r.Congestion < 1 || r.Congestion > 10 {
			return fmt.Errorf("road %d congestion must be 1..10", r.ID)
		}
	}

	for _, ev := range s.Events {
		if ev.Tick < 1 {
			return fmt.Errorf("event tick must be >= 1, got %d", ev.Tick)
		}
		if ev.Type != "emergency" {
			return fmt.Errorf("unsupported event type %q", ev.Type)
		}
		if ev.Plate == "" {
			return errors.New("event.plate is required")
		}
		if _, ok := byID[ev.From]; !ok {
			return fmt.Errorf("event.from intersection %d not found", ev.From)
		}
		if _, ok := byID[ev.To]; !ok {
			return fmt.Errorf("event.to intersection %d not found", ev.To)
		}
	}

	return nil
}

// Built holds the result of building a Scenario: a Graph and intersection metadata.
type Built struct {
	Graph         *model.Graph
	Intersections map[int]Intersection
}

// Build constructs a model.Graph from the scenario's roads and intersections.
// IDs must be contiguous 0..N-1.
func (s *Scenario) Build() (*Built, error) {
	ids := make([]int, 0, len(s.Intersections))
	byID := make(map[int]Intersection, len(s.Intersections))
	for _, in := range s.Intersections {
		ids = append(ids, in.ID)
		byID[in.ID] = in
	}
	sort.Ints(ids)
	for i := 0; i < len(ids); i++ {
		if ids[i] != i {
			return nil, fmt.Errorf("intersection ids must be contiguous 0..N-1 (missing %d)", i)
		}
	}

	g := model.NewGraph(len(s.Intersections))
	for _, r := range s.Roads {
		road := &model.Road{
			ID: r.ID, From: r.From, To: r.To,
			Distance: r.DistanceKm, SpeedLimit: r.SpeedKmph, Congestion: r.Congestion,
		}
		g.AddEdge(r.From, r.To, road)
	}
	return &Built{Graph: g, Intersections: byID}, nil
}

// XY returns the coordinates of an intersection by ID.
func (b *Built) XY(id int) (x, y int, ok bool) {
	in, ok := b.Intersections[id]
	if !ok || in.X == nil || in.Y == nil {
		return 0, 0, false
	}
	return *in.X, *in.Y, true
}
