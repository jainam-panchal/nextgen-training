package dijkstra

import (
	"math"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/ds"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
)

// WeightMinutes returns the travel time cost for traversing a road, in minutes.
//
// Formula: (distance / speed) * (1 + (congestion-1)/10) * 60
//
// Congestion acts as a multiplier:
//   - congestion=1  → multiply by 1.0 (free flow)
//   - congestion=5  → multiply by 1.4 (moderate)
//   - congestion=10 → multiply by 1.9 (heavy)
func WeightMinutes(r *model.Road) float64 {
	if r.SpeedLimit <= 0 {
		return r.Distance
	}
	base := r.Distance / r.SpeedLimit
	mult := 1.0 + float64(r.Congestion-1)/10.0
	return base * mult * 60.0
}

// ShortestPath runs Dijkstra's algorithm on the graph from src to dst.
// Returns the path as a slice of node IDs and the total travel time in minutes.
// Returns (nil, +Inf) if dst is unreachable.
func ShortestPath(g model.GraphReader, src, dst int) ([]int, float64) {
	n := g.Nodes()
	dist := make([]float64, n)
	prev := make([]int, n)
	for i := 0; i < n; i++ {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}

	pq := ds.NewPriorityQueue[int]()

	dist[src] = 0
	pq.PushItem(src, 0)

	for {
		u, d, ok := pq.PopItem()
		if !ok {
			break
		}
		if d > dist[u] {
			continue
		}
		if u == dst {
			break
		}
		for _, e := range g.Neighbors(u) {
			v := e.To
			alt := dist[u] + WeightMinutes(e.Road)
			if alt < dist[v] {
				dist[v] = alt
				prev[v] = u
				pq.PushItem(v, alt)
			}
		}
	}

	if math.IsInf(dist[dst], 1) {
		return nil, math.Inf(1)
	}

	// Reconstruct path from dst back to src.
	var path []int
	for u := dst; u != -1; u = prev[u] {
		path = append([]int{u}, path...)
	}

	return path, dist[dst]
}
