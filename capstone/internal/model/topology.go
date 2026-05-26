package model

import "math"

// Road is a directed edge in the traffic graph.
type Road struct {
	ID         int
	From       int     // start intersection ID
	To         int     // end intersection ID
	Distance   float64 // length in kilometres
	SpeedLimit float64 // max speed in km/h
	Congestion int     // congestion multiplier 1..10 (1 = free flow, 10 = gridlock)
}

// Edge connects one node (From, implicit in the adjacency map) to another.
type Edge struct {
	Road *Road // the road segment (nil for placeholder entries)
	To   int   // destination intersection ID
}

// Graph is a directed adjacency-list graph of intersections and roads.
type Graph struct {
	Nodes int            // number of intersections
	Adj   map[int][]Edge // adjacency list: Adj[from] = list of outgoing edges
}

// NewGraph creates a graph with n nodes and initialised adjacency map.
func NewGraph(n int) *Graph {
	return &Graph{Nodes: n, Adj: make(map[int][]Edge)}
}

// AddEdge adds a directed edge from→to with the given road.
func (g *Graph) AddEdge(from, to int, road *Road) {
	g.Adj[from] = append(g.Adj[from], Edge{Road: road, To: to})
}

// GenerateDemoGraph builds a 20-node chain 0-1-2-...-19 with bidirectional
// roads between consecutive nodes plus a few long shortcuts.
//
//	Chain: 0 ↔ 1 ↔ 2 ↔ ... ↔ 19
//	Shortcuts: 0↔5, 5↔10, 10↔15, 15↔19, 2↔12, 7↔17
//
// Shortcut IDs: 1000-1005 (forward), 1100-1105 (reverse).
func GenerateDemoGraph() *Graph {
	g := NewGraph(20)
	// Base chain: 0 <-> 1 <-> ... <-> 19
	for i := 0; i < 19; i++ {
		road := &Road{ID: i, From: i, To: i + 1, Distance: 1.0, SpeedLimit: 30.0, Congestion: 1}
		g.AddEdge(i, i+1, road)
		back := &Road{ID: i + 100, From: i + 1, To: i, Distance: 1.0, SpeedLimit: 30.0, Congestion: 1}
		g.AddEdge(i+1, i, back)
	}

	// A few shortcut roads to make routing non-trivial.
	shortcuts := [][4]float64{
		// from, to, distanceKm, speedKmph
		{0, 5, 3.2, 50},
		{5, 10, 3.0, 50},
		{10, 15, 3.1, 50},
		{15, 19, 2.6, 50},
		{2, 12, 6.5, 60},
		{7, 17, 6.0, 60},
	}
	baseID := 1000
	for i, sc := range shortcuts {
		from := int(sc[0])
		to := int(sc[1])
		d := math.Round(sc[2]*10) / 10
		s := sc[3]
		g.AddEdge(from, to, &Road{ID: baseID + i, From: from, To: to, Distance: d, SpeedLimit: s, Congestion: 1})
		g.AddEdge(to, from, &Road{ID: baseID + 100 + i, From: to, To: from, Distance: d, SpeedLimit: s, Congestion: 1})
	}
	return g
}

// GraphReader is the interface Dijkstra uses to read the graph.
// The engine implements this via graphView which injects live congestion.
type GraphReader interface {
	Nodes() int
	Neighbors(node int) []Edge
	GetRoad(roadID int) (*Road, bool)
}

// RoadUpdater is the interface for updating congestion values.
// Currently unused by the engine (congestion updates are internal).
type RoadUpdater interface {
	UpdateCongestion(roadID int, delta int) error
	SetCongestion(roadID int, v int) error
}
