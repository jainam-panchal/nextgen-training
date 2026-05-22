package model

type Road struct {
	ID         int
	From       int
	To         int
	Distance   float64
	SpeedLimit float64
	Congestion int // 1..10
}

type Edge struct {
	Road *Road
	To   int
}

type Graph struct {
	Nodes int
	Adj   map[int][]Edge
}

func NewGraph(n int) *Graph {
	return &Graph{Nodes: n, Adj: make(map[int][]Edge)}
}

func (g *Graph) AddEdge(from, to int, road *Road) {
	g.Adj[from] = append(g.Adj[from], Edge{Road: road, To: to})
}

func GenerateDemoGraph() *Graph {
	g := NewGraph(20)
	for i := 0; i < 19; i++ {
		road := &Road{ID: i, From: i, To: i + 1, Distance: 1.0, SpeedLimit: 30.0, Congestion: 1}
		g.AddEdge(i, i+1, road)
		back := &Road{ID: i + 100, From: i + 1, To: i, Distance: 1.0, SpeedLimit: 30.0, Congestion: 1}
		g.AddEdge(i+1, i, back)
	}
	return g
}

type GraphReader interface {
	Nodes() int
	Neighbors(node int) []Edge
	GetRoad(roadID int) (*Road, bool)
}

type RoadUpdater interface {
	UpdateCongestion(roadID int, delta int) error
	SetCongestion(roadID int, v int) error
}
