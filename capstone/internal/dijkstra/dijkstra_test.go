package dijkstra

import (
	"math"
	"testing"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
)

type simpleGraphReader struct {
	base *model.Graph
}

func (s *simpleGraphReader) Nodes() int { return s.base.Nodes }

func (s *simpleGraphReader) Neighbors(node int) []model.Edge {
	return s.base.Adj[node]
}

func (s *simpleGraphReader) GetRoad(roadID int) (*model.Road, bool) {
	for _, edges := range s.base.Adj {
		for _, e := range edges {
			if e.Road != nil && e.Road.ID == roadID {
				return e.Road, true
			}
		}
	}
	return nil, false
}

func TestShortestPath_DemoGraph(t *testing.T) {
	g := &simpleGraphReader{base: model.GenerateDemoGraph()}
	path, travel := ShortestPath(g, 0, 3)
	exp := []int{0, 1, 2, 3}
	if len(path) != len(exp) {
		t.Fatalf("expected path %v, got %v", exp, path)
	}
	for i := range exp {
		if path[i] != exp[i] {
			t.Fatalf("path mismatch at %d: expected %d got %d", i, exp[i], path[i])
		}
	}
	expTime := 3.0 * (1.0 / 30.0) * 60.0
	if math.Abs(travel-expTime) > 1e-9 {
		t.Fatalf("expected travel time %v, got %v", expTime, travel)
	}
}

func TestShortestPath_SameNode(t *testing.T) {
	g := &simpleGraphReader{base: model.GenerateDemoGraph()}
	path, travel := ShortestPath(g, 5, 5)
	if len(path) != 1 || path[0] != 5 {
		t.Fatalf("expected trivial path [5], got %v", path)
	}
	if travel != 0 {
		t.Fatalf("expected travel 0, got %v", travel)
	}
}
