package heap

import (
	"math/rand"
	"strconv"
	"testing"
)

func benchmarkInput(size int) []int {
	values := make([]int, size)
	for i := range values {
		values[i] = rand.Intn(size * 10)
	}
	return values
}

func BenchmarkMinHeapPush(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		input := benchmarkInput(size)
		b.Run("n="+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				h := NewMinHeap(func(a, c int) bool { return a < c })
				for _, value := range input {
					h.Push(value)
				}
			}
		})
	}
}

func BenchmarkMinHeapPop(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		input := benchmarkInput(size)
		b.Run("n="+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				h := NewMinHeap(func(a, c int) bool { return a < c })
				for _, value := range input {
					h.Push(value)
				}
				for !h.IsEmpty() {
					_, _ = h.Pop()
				}
			}
		})
	}
}

func BenchmarkMinHeapMixedPushPop(b *testing.B) {
	for _, size := range []int{1000, 10000} {
		input := benchmarkInput(size)
		b.Run("n="+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				h := NewMinHeap(func(a, c int) bool { return a < c })
				for idx, value := range input {
					h.Push(value)
					if idx%2 == 0 {
						_, _ = h.Pop()
					}
				}
			}
		})
	}
}

// 10K memory-focused benchmark for profile capture.
func BenchmarkMinHeapInsert10K(b *testing.B) {
	input := benchmarkInput(10000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h := NewMinHeap(func(a, c int) bool { return a < c })
		for _, value := range input {
			h.Push(value)
		}
	}
}
