package model

import (
	"errors"
	"sync"
)

type ThreadSafeGraph struct {
	g        *Graph
	roadLock map[int]*sync.Mutex
	roadMap  map[int]*Road
	mu       sync.RWMutex
}

func NewThreadSafeGraph(g *Graph) *ThreadSafeGraph {
	t := &ThreadSafeGraph{
		g:        g,
		roadLock: make(map[int]*sync.Mutex),
		roadMap:  make(map[int]*Road),
	}
	for _, edges := range g.Adj {
		for i := range edges {
			r := edges[i].Road
			t.roadMap[r.ID] = r
			if _, ok := t.roadLock[r.ID]; !ok {
				t.roadLock[r.ID] = &sync.Mutex{}
			}
		}
	}
	return t
}

func (t *ThreadSafeGraph) Nodes() int {
	return t.g.Nodes
}

func (t *ThreadSafeGraph) Neighbors(node int) []Edge {
	t.mu.RLock()
	defer t.mu.RUnlock()
	edges := t.g.Adj[node]
	out := make([]Edge, len(edges))
	copy(out, edges)
	return out
}

// Caller should not mutate Road fields directly; use RoadUpdater methods.
func (t *ThreadSafeGraph) GetRoad(roadID int) (*Road, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	r, ok := t.roadMap[roadID]
	return r, ok
}

func (t *ThreadSafeGraph) UpdateCongestion(roadID int, delta int) error {
	lock, ok := t.roadLock[roadID]
	if !ok {
		return errors.New("road not found")
	}
	lock.Lock()
	defer lock.Unlock()
	r := t.roadMap[roadID]
	r.Congestion += delta
	if r.Congestion < 1 {
		r.Congestion = 1
	}
	if r.Congestion > 10 {
		r.Congestion = 10
	}
	return nil
}

func (t *ThreadSafeGraph) SetCongestion(roadID int, v int) error {
	if v < 1 || v > 10 {
		return errors.New("congestion out of range")
	}
	lock, ok := t.roadLock[roadID]
	if !ok {
		return errors.New("road not found")
	}
	lock.Lock()
	defer lock.Unlock()
	t.roadMap[roadID].Congestion = v
	return nil
}
