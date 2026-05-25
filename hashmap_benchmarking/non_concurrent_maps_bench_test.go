package main

import (
	"fmt"
	"strings"
	"testing"
)

var sink int

// var scales = []int{1_000, 10_000, 100_000, 1_000_000}
// var keySizes = []int{16, 64, 256}
// var shardCounts = []int{1, 16}

var (
	scales      = []int{100_000}
	keySizes    = []int{32}
	shardCounts = []int{1, 16}
)

const keyPrefix = "key-"

func makeKeys(count int, size int, prefix string) []string {
	const indexWidth = 9
	if size < len(prefix)+indexWidth {
		panic("key size too small")
	}

	keys := make([]string, count)
	for i := range keys {
		key := fmt.Sprintf("%s%09d", prefix, i)
		keys[i] = key + strings.Repeat("x", size-len(key))
	}
	return keys
}

func BenchmarkNonConcurrentBuild(b *testing.B) {
	for _, n := range scales {
		for _, keySize := range keySizes {
			keys := makeKeys(n, keySize, keyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					m := NewBuiltinMap()
					for idx, k := range keys {
						m.Set(k, idx)
					}
					sink += m.Len()
				}
			})

			b.Run(fmt.Sprintf("SyncMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					m := NewSyncMap()
					for idx, k := range keys {
						m.Set(k, idx)
					}
					sink += m.Len()
				}
			})

			for _, shards := range shardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shards), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						m := NewShardedBuiltinMap(shards)
						for idx, k := range keys {
							m.Set(k, idx)
						}
						sink += m.Len()
					}
				})
			}
		}
	}
}

func BenchmarkNonConcurrentRead(b *testing.B) {
	for _, n := range scales {
		for _, keySize := range keySizes {
			keys := makeKeys(n, keySize, keyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				m := NewBuiltinMap()
				for idx, k := range keys {
					m.Set(k, idx)
				}
				b.ReportAllocs()
				keyIndex := 0
				for i := 0; i < b.N; i++ {
					v, _ := m.Get(keys[keyIndex])
					sink += v
					keyIndex++
					if keyIndex >= len(keys) {
						keyIndex = 0
					}
				}
			})

			b.Run(fmt.Sprintf("SyncMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				m := NewSyncMap()
				for idx, k := range keys {
					m.Set(k, idx)
				}
				b.ReportAllocs()
				keyIndex := 0
				for i := 0; i < b.N; i++ {
					v, _ := m.Get(keys[keyIndex])
					sink += v
					keyIndex++
					if keyIndex >= len(keys) {
						keyIndex = 0
					}
				}
			})

			for _, shards := range shardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shards), func(b *testing.B) {
					m := NewShardedBuiltinMap(shards)
					for idx, k := range keys {
						m.Set(k, idx)
					}
					b.ReportAllocs()
					keyIndex := 0
					for i := 0; i < b.N; i++ {
						v, _ := m.Get(keys[keyIndex])
						sink += v
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
	for _, n := range scales {
		for _, keySize := range keySizes {
			keys := makeKeys(n, keySize, keyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				m := NewBuiltinMap()
				for idx, k := range keys {
					m.Set(k, idx)
				}
				b.ReportAllocs()
				keyIndex := 0
				for i := 0; i < b.N; i++ {
					m.Set(keys[keyIndex], i)
					keyIndex++
					if keyIndex >= len(keys) {
						keyIndex = 0
					}
				}
				sink += m.Len()
			})

			b.Run(fmt.Sprintf("SyncMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				m := NewSyncMap()
				for idx, k := range keys {
					m.Set(k, idx)
				}
				b.ReportAllocs()
				keyIndex := 0
				for i := 0; i < b.N; i++ {
					m.Set(keys[keyIndex], i)
					keyIndex++
					if keyIndex >= len(keys) {
						keyIndex = 0
					}
				}
				sink += m.Len()
			})

			for _, shards := range shardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shards), func(b *testing.B) {
					m := NewShardedBuiltinMap(shards)
					for idx, k := range keys {
						m.Set(k, idx)
					}
					b.ReportAllocs()
					keyIndex := 0
					for i := 0; i < b.N; i++ {
						m.Set(keys[keyIndex], i)
						keyIndex++
						if keyIndex >= len(keys) {
							keyIndex = 0
						}
					}
					sink += m.Len()
				})
			}
		}
	}
}

func BenchmarkNonConcurrentDeleteOnly(b *testing.B) {
	for _, n := range scales {
		for _, keySize := range keySizes {
			keys := makeKeys(n, keySize, keyPrefix)

			b.Run(fmt.Sprintf("BuiltinMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					m := NewBuiltinMap()
					for idx, k := range keys {
						m.Set(k, idx)
					}
					b.StartTimer()

					for _, k := range keys {
						m.Delete(k)
					}
					sink += m.Len()
				}
			})

			b.Run(fmt.Sprintf("SyncMap/n=%d/keysize=%d", n, keySize), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					m := NewSyncMap()
					for idx, k := range keys {
						m.Set(k, idx)
					}
					b.StartTimer()

					for _, k := range keys {
						m.Delete(k)
					}
					sink += m.Len()
				}
			})

			for _, shards := range shardCounts {
				b.Run(fmt.Sprintf("ShardedBuiltinMap/n=%d/keysize=%d/shards=%d", n, keySize, shards), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						m := NewShardedBuiltinMap(shards)
						for idx, k := range keys {
							m.Set(k, idx)
						}
						b.StartTimer()

						for _, k := range keys {
							m.Delete(k)
						}
						sink += m.Len()
					}
				})
			}
		}
	}
}
