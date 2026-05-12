package main

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

var benchmarkValue atomic.Int64
var benchmarkFound atomic.Bool

const (
	keyPrefix    = "key-"
	keyCount     = 1_00_00_000
	keySizeBytes = 25
	shardCount   = 32
)

func makeBenchmarkKeys(count int, keySizeBytes int, prefix string) []string {
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

func BenchmarkBuiltinMapBuild(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	b.ReportAllocs()
	b.ReportMetric(float64(keyCount), "keys/op")
	b.ResetTimer()

	for iteration := 0; iteration < b.N; iteration++ {
		hashMap := NewBuiltinMap()

		for keyIndex, key := range keys {
			hashMap.Set(key, keyIndex)
		}
	}
}

func BenchmarkShardedBuiltinMapBuild(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	b.ReportAllocs()
	b.ReportMetric(float64(keyCount), "keys/op")
	b.ResetTimer()

	for iteration := 0; iteration < b.N; iteration++ {
		shardedMap := NewShardedBuiltinMap(shardCount)

		for keyIndex, key := range keys {
			shardedMap.Set(key, keyIndex)
		}
	}
}

func BenchmarkBuiltinMapGet(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	hashMap := NewBuiltinMap()
	for keyIndex, key := range keys {
		hashMap.Set(key, keyIndex)
	}

	var value int
	var found bool

	b.ReportAllocs()
	b.ResetTimer()

	keyIndex := 0

	for iteration := 0; iteration < b.N; iteration++ {
		value, found = hashMap.Get(keys[keyIndex])

		keyIndex++
		if keyIndex == len(keys) {
			keyIndex = 0
		}
	}

	benchmarkValue.Store(int64(value))
	benchmarkFound.Store(found)
}

func BenchmarkShardedBuiltinMapGet(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	shardedMap := NewShardedBuiltinMap(shardCount)
	for keyIndex, key := range keys {
		shardedMap.Set(key, keyIndex)
	}

	var value int
	var found bool

	b.ReportAllocs()
	b.ResetTimer()

	keyIndex := 0

	for iteration := 0; iteration < b.N; iteration++ {
		value, found = shardedMap.Get(keys[keyIndex])

		keyIndex++
		if keyIndex == len(keys) {
			keyIndex = 0
		}
	}

	benchmarkValue.Store(int64(value))
	benchmarkFound.Store(found)
}

func BenchmarkBuiltinMapDelete(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	b.ReportAllocs()
	b.ReportMetric(float64(keyCount), "keys/op")

	for iteration := 0; iteration < b.N; iteration++ {
		b.StopTimer()

		hashMap := NewBuiltinMap()
		for keyIndex, key := range keys {
			hashMap.Set(key, keyIndex)
		}

		b.StartTimer()

		for _, key := range keys {
			hashMap.Delete(key)
		}
	}
}

func BenchmarkShardedBuiltinMapDelete(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	b.ReportAllocs()
	b.ReportMetric(float64(keyCount), "keys/op")

	for iteration := 0; iteration < b.N; iteration++ {
		b.StopTimer()

		shardedMap := NewShardedBuiltinMap(shardCount)
		for keyIndex, key := range keys {
			shardedMap.Set(key, keyIndex)
		}

		b.StartTimer()

		for _, key := range keys {
			shardedMap.Delete(key)
		}
	}
}

// parallel ------------

func BenchmarkShardedBuiltinMapParallelGet(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	shardedMap := NewShardedBuiltinMap(shardCount)
	for keyIndex, key := range keys {
		shardedMap.Set(key, keyIndex)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		keyIndex := 0
		localValue := 0
		localFound := false

		for pb.Next() {
			localValue, localFound = shardedMap.Get(keys[keyIndex])

			keyIndex++
			if keyIndex == len(keys) {
				keyIndex = 0
			}
		}

		benchmarkValue.Store(int64(localValue))
		benchmarkFound.Store(localFound)
	})
}

func BenchmarkBuiltinMapParallelGet(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	hashMap := NewBuiltinMap()
	for keyIndex, key := range keys {
		hashMap.Set(key, keyIndex)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		keyIndex := 0
		localValue := 0
		localFound := false

		for pb.Next() {
			localValue, localFound = hashMap.Get(keys[keyIndex])

			keyIndex++
			if keyIndex == len(keys) {
				keyIndex = 0
			}
		}

		benchmarkValue.Store(int64(localValue))
		benchmarkFound.Store(localFound)
	})
}

func BenchmarkBuiltinMapParallelSet(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	hashMap := NewBuiltinMap()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		keyIndex := 0

		for pb.Next() {
			hashMap.Set(keys[keyIndex], keyIndex)

			keyIndex++
			if keyIndex == len(keys) {
				keyIndex = 0
			}
		}
	})
}

func BenchmarkShardedBuiltinMapParallelSet(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	shardedMap := NewShardedBuiltinMap(shardCount)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		keyIndex := 0

		for pb.Next() {
			shardedMap.Set(keys[keyIndex], keyIndex)

			keyIndex++
			if keyIndex == len(keys) {
				keyIndex = 0
			}
		}
	})
}

func BenchmarkBuiltinMapParallelDelete(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	hashMap := NewBuiltinMap()
	for keyIndex, key := range keys {
		hashMap.Set(key, keyIndex)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		keyIndex := 0

		for pb.Next() {
			hashMap.Delete(keys[keyIndex])

			keyIndex++
			if keyIndex == len(keys) {
				keyIndex = 0
			}
		}
	})
}

func BenchmarkShardedBuiltinMapParallelDelete(b *testing.B) {
	keys := makeBenchmarkKeys(keyCount, keySizeBytes, keyPrefix)

	shardedMap := NewShardedBuiltinMap(shardCount)
	for keyIndex, key := range keys {
		shardedMap.Set(key, keyIndex)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		keyIndex := 0

		for pb.Next() {
			shardedMap.Delete(keys[keyIndex])

			keyIndex++
			if keyIndex == len(keys) {
				keyIndex = 0
			}
		}
	})
}
