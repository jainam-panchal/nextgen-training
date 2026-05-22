package dijkstra

import (
	"math"
	"time"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/ds"
	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
)

func weight(r *model.Road) float64 {
	if r.SpeedLimit <= 0 {
		return r.Distance
	}
	base := r.Distance / r.SpeedLimit
	mult := 1.0 + float64(r.Congestion-1)/10.0
	return base * mult
}

func ShortestPath(g model.GraphReader, src, dst int) ([]int, float64) {
	n := g.Nodes()
	dist := make([]float64, n)
	prev := make([]int, n)
	for i := 0; i < n; i++ {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}

	pq := ds.NewPriorityQueue[int]()
	pq.Init()

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
			alt := dist[u] + weight(e.Road)
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

	var path []int
	for u := dst; u != -1; u = prev[u] {
		path = append([]int{u}, path...)
	}

	_ = time.Second

	return path, dist[dst]
}
