package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
)

var sink int
var iterSink uint32

type densityConfig struct {
	name   string
	factor uint64
}

var densities = []densityConfig{
	{name: "dense", factor: 1},
	{name: "moderate", factor: 10},
	{name: "sparse", factor: 1_000},
	{name: "extreme", factor: 0},
}

type distConfig struct {
	name       string
	contiguous bool
}

var dists = []distConfig{
	{name: "contiguous", contiguous: true},
	{name: "random", contiguous: false},
}

var ns = []int{1_000, 10_000, 100_000, 1_000_000}

func rangeSize(n int, d densityConfig) uint64 {
	if d.factor == 0 {
		return math.MaxUint32
	}
	r := uint64(n) * d.factor
	if r > math.MaxUint32 {
		return math.MaxUint32
	}
	return r
}

func bitArrayUsable(n int, d densityConfig) bool {
	rs := rangeSize(n, d)
	bitsetBytes := uint64(rs) / 8
	return bitsetBytes <= 64*1024*1024
}

func makeKeys(n int, d densityConfig, contiguous bool) []uint32 {
	rng := rand.New(rand.NewPCG(42, 42))
	rs := rangeSize(n, d)
	keys := make([]uint32, n)

	if contiguous {
		offset := uint64(rng.Uint32()) % (rs - uint64(n) + 1)
		for i := range keys {
			keys[i] = uint32(offset + uint64(i))
		}
		return keys
	}

	if rs < uint64(n)*4 {
		perm := rng.Perm(int(rs))
		for i := range keys {
			keys[i] = uint32(perm[i])
		}
	} else {
		seen := make(map[uint32]struct{}, n)
		for len(seen) < n {
			x := rng.Uint32() % uint32(rs)
			if _, ok := seen[x]; !ok {
				seen[x] = struct{}{}
				keys[len(seen)-1] = x
			}
		}
	}
	return keys
}

func makeOverlappingSets(n int, d densityConfig, dist distConfig) (a, b, shared, onlyA, onlyB []uint32) {
	half := n / 2
	shared = makeKeys(half, d, dist.contiguous)

	rangeA := make([]uint32, half)
	rangeB := make([]uint32, half)
	onlyA = rangeA
	onlyB = rangeB

	offsetA := uint64(n)
	offsetB := uint64(n)
	rs := rangeSize(n, d)

	if rs > uint64(n)*2 {
		if dist.contiguous {
			extraStart := rs / 2
			for i := range onlyA {
				onlyA[i] = uint32(extraStart + uint64(i))
			}
			for i := range onlyB {
				onlyB[i] = uint32(extraStart + uint64(n/2) + uint64(i))
			}
		} else {
			taken := make(map[uint32]struct{}, n+half)
			for _, x := range shared {
				taken[x] = struct{}{}
			}
			rng := rand.New(rand.NewPCG(99, 99))
			for i := range onlyA {
				for {
					x := rng.Uint32() % uint32(rs)
					if _, ok := taken[x]; !ok {
						taken[x] = struct{}{}
						onlyA[i] = x
						break
					}
				}
			}
			for i := range onlyB {
				for {
					x := rng.Uint32() % uint32(rs)
					if _, ok := taken[x]; !ok {
						taken[x] = struct{}{}
						onlyB[i] = x
						break
					}
				}
			}
		}
	} else {
		for i := range onlyA {
			onlyA[i] = uint32(offsetA + uint64(i))
		}
		for i := range onlyB {
			onlyB[i] = uint32(offsetB + uint64(i))
		}
	}

	return shared, onlyA, onlyB, shared, shared
}

func BenchmarkBuild(b *testing.B) {
	for _, n := range ns {
		for _, d := range densities {
			for _, dist := range dists {
				keys := makeKeys(n, d, dist.contiguous)

				b.Run(fmt.Sprintf("GoSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						s := NewGoSet()
						for _, k := range keys {
							s.Add(k)
						}
						sink += int(s.Len())
					}
				})

				b.Run(fmt.Sprintf("RoaringSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						s := NewRoaringSet()
						for _, k := range keys {
							s.Add(k)
						}
						sink += int(s.Len())
					}
				})

				b.Run(fmt.Sprintf("BitArraySet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					if !bitArrayUsable(n, d) {
						b.Skip("bitset too large")
					}
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						s := NewBitArraySet()
						for _, k := range keys {
							s.Add(k)
						}
						sink += int(s.Len())
					}
				})
			}
		}
	}
}

func BenchmarkLookupHit(b *testing.B) {
	for _, n := range ns {
		for _, d := range densities {
			for _, dist := range dists {
				keys := makeKeys(n, d, dist.contiguous)

				b.Run(fmt.Sprintf("GoSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					s := NewGoSet()
					for _, k := range keys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					keyIndex := 0
					for i := 0; i < b.N; i++ {
						if s.Contains(keys[keyIndex]) {
							sink++
						}
						keyIndex++
						if keyIndex >= len(keys) {
							keyIndex = 0
						}
					}
				})

				b.Run(fmt.Sprintf("RoaringSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					s := NewRoaringSet()
					for _, k := range keys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					keyIndex := 0
					for i := 0; i < b.N; i++ {
						if s.Contains(keys[keyIndex]) {
							sink++
						}
						keyIndex++
						if keyIndex >= len(keys) {
							keyIndex = 0
						}
					}
				})

				b.Run(fmt.Sprintf("BitArraySet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					if !bitArrayUsable(n, d) {
						b.Skip("bitset too large")
					}
					s := NewBitArraySet()
					for _, k := range keys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					keyIndex := 0
					for i := 0; i < b.N; i++ {
						if s.Contains(keys[keyIndex]) {
							sink++
						}
						keyIndex++
						if keyIndex >= len(keys) {
							keyIndex = 0
						}
					}
				})
			}
		}
	}
}

func BenchmarkLookupMiss(b *testing.B) {
	for _, n := range ns {
		for _, d := range densities {
			for _, dist := range dists {
				memberKeys := makeKeys(n, d, dist.contiguous)

				missKeys := make([]uint32, n)
				rs := rangeSize(n, d)
				missRS := rs + uint64(n)
				if missRS > math.MaxUint32 {
					missRS = math.MaxUint32
				}
				rng := rand.New(rand.NewPCG(77, 77))
				member := make(map[uint32]struct{}, n)
				for _, k := range memberKeys {
					member[k] = struct{}{}
				}
				for i := range missKeys {
					for {
						x := rng.Uint32() % uint32(missRS)
						if _, ok := member[x]; !ok {
							missKeys[i] = x
							break
						}
					}
				}

				b.Run(fmt.Sprintf("GoSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					s := NewGoSet()
					for _, k := range memberKeys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					keyIndex := 0
					for i := 0; i < b.N; i++ {
						if s.Contains(missKeys[keyIndex]) {
							sink++
						}
						keyIndex++
						if keyIndex >= len(missKeys) {
							keyIndex = 0
						}
					}
				})

				b.Run(fmt.Sprintf("RoaringSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					s := NewRoaringSet()
					for _, k := range memberKeys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					keyIndex := 0
					for i := 0; i < b.N; i++ {
						if s.Contains(missKeys[keyIndex]) {
							sink++
						}
						keyIndex++
						if keyIndex >= len(missKeys) {
							keyIndex = 0
						}
					}
				})

				b.Run(fmt.Sprintf("BitArraySet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					if !bitArrayUsable(n, d) {
						b.Skip("bitset too large")
					}
					s := NewBitArraySet()
					for _, k := range memberKeys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					keyIndex := 0
					for i := 0; i < b.N; i++ {
						if s.Contains(missKeys[keyIndex]) {
							sink++
						}
						keyIndex++
						if keyIndex >= len(missKeys) {
							keyIndex = 0
						}
					}
				})
			}
		}
	}
}

func BenchmarkIterate(b *testing.B) {
	for _, n := range ns {
		for _, d := range densities {
			for _, dist := range dists {
				keys := makeKeys(n, d, dist.contiguous)

				b.Run(fmt.Sprintf("GoSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					s := NewGoSet()
					for _, k := range keys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						s.Iterate(func(x uint32) bool {
							iterSink += x
							return true
						})
					}
				})

				b.Run(fmt.Sprintf("RoaringSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					s := NewRoaringSet()
					for _, k := range keys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						s.Iterate(func(x uint32) bool {
							iterSink += x
							return true
						})
					}
				})

				b.Run(fmt.Sprintf("BitArraySet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					if !bitArrayUsable(n, d) {
						b.Skip("bitset too large")
					}
					s := NewBitArraySet()
					for _, k := range keys {
						s.Add(k)
					}
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						s.Iterate(func(x uint32) bool {
							iterSink += x
							return true
						})
					}
				})
			}
		}
	}
}

func BenchmarkUnion(b *testing.B) {
	for _, n := range ns {
		for _, d := range densities {
			for _, dist := range dists {
				shared, onlyA, onlyB, _, _ := makeOverlappingSets(n, d, dist)

				goA := NewGoSet()
				for _, k := range shared {
					goA.Add(k)
				}
				for _, k := range onlyA {
					goA.Add(k)
				}
				goB := NewGoSet()
				for _, k := range shared {
					goB.Add(k)
				}
				for _, k := range onlyB {
					goB.Add(k)
				}

				rbA := NewRoaringSet()
				for _, k := range shared {
					rbA.Add(k)
				}
				for _, k := range onlyA {
					rbA.Add(k)
				}
				rbB := NewRoaringSet()
				for _, k := range shared {
					rbB.Add(k)
				}
				for _, k := range onlyB {
					rbB.Add(k)
				}

				baA := NewBitArraySet()
				for _, k := range shared {
					baA.Add(k)
				}
				for _, k := range onlyA {
					baA.Add(k)
				}
				baB := NewBitArraySet()
				for _, k := range shared {
					baB.Add(k)
				}
				for _, k := range onlyB {
					baB.Add(k)
				}

				b.Run(fmt.Sprintf("GoSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						u := make(map[uint32]struct{}, len(goA.m)+len(goB.m))
						for k := range goA.m {
							u[k] = struct{}{}
						}
						for k := range goB.m {
							u[k] = struct{}{}
						}
						sink += len(u)
					}
				})

				b.Run(fmt.Sprintf("RoaringSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						u := roaring.Or(rbA.b, rbB.b)
						sink += int(u.GetCardinality())
					}
				})

				b.Run(fmt.Sprintf("BitArraySet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					if !bitArrayUsable(n, d) {
						b.Skip("bitset too large")
					}
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						u := bitArrayUnion(baA, baB)
						sink += int(u.Len())
					}
				})
			}
		}
	}
}

func BenchmarkIntersection(b *testing.B) {
	for _, n := range ns {
		for _, d := range densities {
			for _, dist := range dists {
				shared, onlyA, onlyB, _, _ := makeOverlappingSets(n, d, dist)

				goA := NewGoSet()
				for _, k := range shared {
					goA.Add(k)
				}
				for _, k := range onlyA {
					goA.Add(k)
				}
				goB := NewGoSet()
				for _, k := range shared {
					goB.Add(k)
				}
				for _, k := range onlyB {
					goB.Add(k)
				}

				rbA := NewRoaringSet()
				for _, k := range shared {
					rbA.Add(k)
				}
				for _, k := range onlyA {
					rbA.Add(k)
				}
				rbB := NewRoaringSet()
				for _, k := range shared {
					rbB.Add(k)
				}
				for _, k := range onlyB {
					rbB.Add(k)
				}

				baA := NewBitArraySet()
				for _, k := range shared {
					baA.Add(k)
				}
				for _, k := range onlyA {
					baA.Add(k)
				}
				baB := NewBitArraySet()
				for _, k := range shared {
					baB.Add(k)
				}
				for _, k := range onlyB {
					baB.Add(k)
				}

				b.Run(fmt.Sprintf("GoSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						result := make(map[uint32]struct{})
						for k := range goA.m {
							if _, ok := goB.m[k]; ok {
								result[k] = struct{}{}
							}
						}
						sink += len(result)
					}
				})

				b.Run(fmt.Sprintf("RoaringSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						inter := roaring.And(rbA.b, rbB.b)
						sink += int(inter.GetCardinality())
					}
				})

				b.Run(fmt.Sprintf("BitArraySet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					if !bitArrayUsable(n, d) {
						b.Skip("bitset too large")
					}
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						inter := bitArrayIntersection(baA, baB)
						sink += int(inter.Len())
					}
				})
			}
		}
	}
}

func BenchmarkDelete(b *testing.B) {
	for _, n := range ns {
		for _, d := range densities {
			for _, dist := range dists {
				keys := makeKeys(n, d, dist.contiguous)

				b.Run(fmt.Sprintf("GoSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						s := NewGoSet()
						for _, k := range keys {
							s.Add(k)
						}
						b.StartTimer()
						for _, k := range keys {
							s.Remove(k)
						}
						sink += int(s.Len())
					}
				})

				b.Run(fmt.Sprintf("RoaringSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						s := NewRoaringSet()
						for _, k := range keys {
							s.Add(k)
						}
						b.StartTimer()
						for _, k := range keys {
							s.Remove(k)
						}
						sink += int(s.Len())
					}
				})

				b.Run(fmt.Sprintf("BitArraySet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					if !bitArrayUsable(n, d) {
						b.Skip("bitset too large")
					}
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						s := NewBitArraySet()
						for _, k := range keys {
							s.Add(k)
						}
						b.StartTimer()
						for _, k := range keys {
							s.Remove(k)
						}
						sink += int(s.Len())
					}
				})
			}
		}
	}
}

func BenchmarkMemory(b *testing.B) {
	for _, n := range ns {
		for _, d := range densities {
			for _, dist := range dists {
				keys := makeKeys(n, d, dist.contiguous)

				b.Run(fmt.Sprintf("GoSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					var totalBytes uint64
					for i := 0; i < b.N; i++ {
						runtime.GC()
						var m1 runtime.MemStats
						runtime.ReadMemStats(&m1)

						s := NewGoSet()
						for _, k := range keys {
							s.Add(k)
						}
						sink += int(s.Len())

						var m2 runtime.MemStats
						runtime.ReadMemStats(&m2)
						totalBytes += m2.Alloc - m1.Alloc
					}
					b.ReportMetric(float64(totalBytes)/float64(b.N), "heapdelta/B")
				})

				b.Run(fmt.Sprintf("RoaringSet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					b.ReportAllocs()
					var totalBytes uint64
					for i := 0; i < b.N; i++ {
						runtime.GC()
						var m1 runtime.MemStats
						runtime.ReadMemStats(&m1)

						s := NewRoaringSet()
						for _, k := range keys {
							s.Add(k)
						}
						sink += int(s.Len())

						var m2 runtime.MemStats
						runtime.ReadMemStats(&m2)
						totalBytes += m2.Alloc - m1.Alloc
					}
					b.ReportMetric(float64(totalBytes)/float64(b.N), "heapdelta/B")
				})

				b.Run(fmt.Sprintf("BitArraySet/%s/%s/n=%d", d.name, dist.name, n), func(b *testing.B) {
					if !bitArrayUsable(n, d) {
						b.Skip("bitset too large")
					}
					b.ReportAllocs()
					var totalBytes uint64
					for i := 0; i < b.N; i++ {
						runtime.GC()
						var m1 runtime.MemStats
						runtime.ReadMemStats(&m1)

						s := NewBitArraySet()
						for _, k := range keys {
							s.Add(k)
						}
						sink += int(s.Len())

						var m2 runtime.MemStats
						runtime.ReadMemStats(&m2)
						totalBytes += m2.Alloc - m1.Alloc
					}
					b.ReportMetric(float64(totalBytes)/float64(b.N), "heapdelta/B")
				})
			}
		}
	}
}
