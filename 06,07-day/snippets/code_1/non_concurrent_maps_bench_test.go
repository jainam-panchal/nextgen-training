package main

import (
	"fmt"
	"strings"
	"testing"
)

// Non-concurrent benchmark suite:
// - Measures latency (ns/op) and memory behavior (B/op, allocs/op) together.
// - Compares BuiltinMap vs ShardedBuiltinMap at multiple scales and key sizes.
// - No locks, no RunParallel, single-threaded benchmark loops.

var nonConcurrentSink int

var nonConcurrentScales = []int{
	1_000,
	10_000,
	100_000,
	1_000_000,
}

var nonConcurrentKeySizes = []int{16, 32, 64, 128, 256}
var nonConcurrentShardCounts = []int{1, 4, 16, 32, 64}

const nonConcurrentKeyPrefix = "key-"

func makeNonConcurrentBenchmarkKeys(count int, keySizeBytes int, prefix string) []string {
	const indexWidth = 9

	if keySizeBytes < len(prefix)+indexWidth {
		panic("keySizeBytes is too small")
	}

	keys := make([]string, count)
	for i := range keys {
		key := fmt.Sprintf("%s%09d", prefix, i)
		keys[i] = key + strings.Repeat("x", keySizeBytes-len(key))
	}

	return keys
}

func BenchmarkNonConcurrentBuild(b *testing.B) {
	for _, n := range nonConcurrentScales {
		for _, keySize := range nonConcurrentKeySizes {
			keys := makeNonConcurrentBenchmarkKeys(n, keySize, nonConcurrentKeyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(n), "keys/op")
				b.ReportMetric(float64(keySize), "key_size_bytes")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					hashMap := NewBuiltinMap()
					for keyIndex, key := range keys {
						hashMap.Set(key, keyIndex)
					}
					nonConcurrentSink += hashMap.Len()
				}
			})

			for _, shardCount := range nonConcurrentShardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shardCount), func(b *testing.B) {
					b.ReportAllocs()
					b.ReportMetric(float64(n), "keys/op")
					b.ReportMetric(float64(keySize), "key_size_bytes")
					b.ReportMetric(float64(shardCount), "shards")
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						hashMap := NewShardedBuiltinMap(shardCount)
						for keyIndex, key := range keys {
							hashMap.Set(key, keyIndex)
						}
						nonConcurrentSink += hashMap.Len()
					}
				})
			}
		}
	}
}

func BenchmarkNonConcurrentRead(b *testing.B) {
	for _, n := range nonConcurrentScales {
		for _, keySize := range nonConcurrentKeySizes {
			keys := makeNonConcurrentBenchmarkKeys(n, keySize, nonConcurrentKeyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				hashMap := NewBuiltinMap()
				for keyIndex, key := range keys {
					hashMap.Set(key, keyIndex)
				}

				b.ReportAllocs()
				b.ReportMetric(float64(n), "keys/op")
				b.ReportMetric(float64(keySize), "key_size_bytes")
				b.ResetTimer()

				keyIndex := 0
				for i := 0; i < b.N; i++ {
					value, _ := hashMap.Get(keys[keyIndex])
					nonConcurrentSink += value
					keyIndex++
					if keyIndex >= len(keys) {
						keyIndex = 0
					}
				}
			})

			for _, shardCount := range nonConcurrentShardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shardCount), func(b *testing.B) {
					hashMap := NewShardedBuiltinMap(shardCount)
					for keyIndex, key := range keys {
						hashMap.Set(key, keyIndex)
					}

					b.ReportAllocs()
					b.ReportMetric(float64(n), "keys/op")
					b.ReportMetric(float64(keySize), "key_size_bytes")
					b.ReportMetric(float64(shardCount), "shards")
					b.ResetTimer()

					keyIndex := 0
					for i := 0; i < b.N; i++ {
						value, _ := hashMap.Get(keys[keyIndex])
						nonConcurrentSink += value
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

func BenchmarkNonConcurrentWriteUpdate(b *testing.B) {
	for _, n := range nonConcurrentScales {
		for _, keySize := range nonConcurrentKeySizes {
			keys := makeNonConcurrentBenchmarkKeys(n, keySize, nonConcurrentKeyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				hashMap := NewBuiltinMap()
				for keyIndex, key := range keys {
					hashMap.Set(key, keyIndex)
				}

				b.ReportAllocs()
				b.ReportMetric(float64(n), "keys/op")
				b.ReportMetric(float64(keySize), "key_size_bytes")
				b.ResetTimer()

				keyIndex := 0
				for i := 0; i < b.N; i++ {
					hashMap.Set(keys[keyIndex], i)
					keyIndex++
					if keyIndex >= len(keys) {
						keyIndex = 0
					}
				}
				nonConcurrentSink += hashMap.Len()
			})

			for _, shardCount := range nonConcurrentShardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shardCount), func(b *testing.B) {
					hashMap := NewShardedBuiltinMap(shardCount)
					for keyIndex, key := range keys {
						hashMap.Set(key, keyIndex)
					}

					b.ReportAllocs()
					b.ReportMetric(float64(n), "keys/op")
					b.ReportMetric(float64(keySize), "key_size_bytes")
					b.ReportMetric(float64(shardCount), "shards")
					b.ResetTimer()

					keyIndex := 0
					for i := 0; i < b.N; i++ {
						hashMap.Set(keys[keyIndex], i)
						keyIndex++
						if keyIndex >= len(keys) {
							keyIndex = 0
						}
					}
					nonConcurrentSink += hashMap.Len()
				})
			}
		}
	}
}

func BenchmarkNonConcurrentDelete(b *testing.B) {
	// Delete benchmark includes setup (build + delete) in timed section.
	// Keep this because end-to-end cost matters for many workflows.
	for _, n := range nonConcurrentScales {
		for _, keySize := range nonConcurrentKeySizes {
			keys := makeNonConcurrentBenchmarkKeys(n, keySize, nonConcurrentKeyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(n), "keys/op")
				b.ReportMetric(float64(keySize), "key_size_bytes")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					hashMap := NewBuiltinMap()
					for keyIndex, key := range keys {
						hashMap.Set(key, keyIndex)
					}
					for _, key := range keys {
						hashMap.Delete(key)
					}
					nonConcurrentSink += hashMap.Len()
				}
			})

			for _, shardCount := range nonConcurrentShardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shardCount), func(b *testing.B) {
					b.ReportAllocs()
					b.ReportMetric(float64(n), "keys/op")
					b.ReportMetric(float64(keySize), "key_size_bytes")
					b.ReportMetric(float64(shardCount), "shards")
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						hashMap := NewShardedBuiltinMap(shardCount)
						for keyIndex, key := range keys {
							hashMap.Set(key, keyIndex)
						}
						for _, key := range keys {
							hashMap.Delete(key)
						}
						nonConcurrentSink += hashMap.Len()
					}
				})
			}
		}
	}
}

func BenchmarkNonConcurrentDeleteOnly(b *testing.B) {
	// Delete-only benchmark excludes setup time from timed section.
	// This isolates pure delete behavior and memory churn from the delete path.
	for _, n := range nonConcurrentScales {
		for _, keySize := range nonConcurrentKeySizes {
			keys := makeNonConcurrentBenchmarkKeys(n, keySize, nonConcurrentKeyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(n), "keys/op")
				b.ReportMetric(float64(keySize), "key_size_bytes")

				for i := 0; i < b.N; i++ {
					b.StopTimer()
					hashMap := NewBuiltinMap()
					for keyIndex, key := range keys {
						hashMap.Set(key, keyIndex)
					}
					b.StartTimer()

					for _, key := range keys {
						hashMap.Delete(key)
					}
					nonConcurrentSink += hashMap.Len()
				}
			})

			for _, shardCount := range nonConcurrentShardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shardCount), func(b *testing.B) {
					b.ReportAllocs()
					b.ReportMetric(float64(n), "keys/op")
					b.ReportMetric(float64(keySize), "key_size_bytes")
					b.ReportMetric(float64(shardCount), "shards")

					for i := 0; i < b.N; i++ {
						b.StopTimer()
						hashMap := NewShardedBuiltinMap(shardCount)
						for keyIndex, key := range keys {
							hashMap.Set(key, keyIndex)
						}
						b.StartTimer()

						for _, key := range keys {
							hashMap.Delete(key)
						}
						nonConcurrentSink += hashMap.Len()
					}
				})
			}
		}
	}
}
