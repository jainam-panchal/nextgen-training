package dijkstra

import (
	"testing"

	"github.com/jainam-panchal/nextgen-training/capstone/internal/model"
)

func BenchmarkShortestPath_20Nodes(b *testing.B) {
	g := &simpleGraphReader{base: model.GenerateDemoGraph()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShortestPath(g, 0, 19)
	}
}

func BenchmarkShortestPath_100Nodes(b *testing.B) {
	g := makeChainGraph(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShortestPath(g, 0, 99)
	}
}

func BenchmarkShortestPath_500Nodes(b *testing.B) {
	g := makeChainGraph(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShortestPath(g, 0, 499)
	}
}

func makeChainGraph(n int) *simpleGraphReader {
	base := model.NewGraph(n)
	for i := 0; i < n-1; i++ {
		base.AddEdge(i, i+1, &model.Road{
			ID: i, From: i, To: i + 1,
			Distance: 1.0, SpeedLimit: 30.0, Congestion: 1,
		})
		base.AddEdge(i+1, i, &model.Road{
			ID: i + 1000, From: i + 1, To: i,
			Distance: 1.0, SpeedLimit: 30.0, Congestion: 1,
		})
	}
	return &simpleGraphReader{base: base}
}
