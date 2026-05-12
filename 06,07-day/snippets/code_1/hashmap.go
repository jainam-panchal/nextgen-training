package main

import (
	"sync"
)

type HashMap interface {
	Set(key string, value int)
	Get(key string) (int, bool)
	Delete(key string)
	Len() int
}

type BuiltinMap struct {
	mutex sync.RWMutex
	items map[string]int
}

func NewBuiltinMap() *BuiltinMap {
	return &BuiltinMap{
		items: make(map[string]int),
	}
}

func (hashMap *BuiltinMap) Set(key string, value int) {
	hashMap.mutex.Lock()
	defer hashMap.mutex.Unlock()

	hashMap.items[key] = value
}

func (hashMap *BuiltinMap) Get(key string) (int, bool) {
	hashMap.mutex.RLock()
	defer hashMap.mutex.RUnlock()

	value, exists := hashMap.items[key]
	return value, exists
}

func (hashMap *BuiltinMap) Delete(key string) {
	hashMap.mutex.Lock()
	defer hashMap.mutex.Unlock()

	delete(hashMap.items, key)
}

func (hashMap *BuiltinMap) Len() int {
	hashMap.mutex.RLock()
	defer hashMap.mutex.RUnlock()

	return len(hashMap.items)
}

type ShardedBuiltinMap struct {
	shards []BuiltinMap
}

func (hashMap *ShardedBuiltinMap) Set(key string, value int) {
	shard := hashMap.getShard(key)
	shard.Set(key, value)
}

func (hashMap *ShardedBuiltinMap) Get(key string) (int, bool) {
	shard := hashMap.getShard(key)
	return shard.Get(key)
}

func (hashMap *ShardedBuiltinMap) Delete(key string) {
	shard := hashMap.getShard(key)
	shard.Delete(key)
}

func (hashMap *ShardedBuiltinMap) Len() int {
	totalLength := 0

	for shardIndex := range hashMap.shards {
		totalLength += hashMap.shards[shardIndex].Len()
	}

	return totalLength
}

func NewShardedBuiltinMap(shardCount int) *ShardedBuiltinMap {
	shards := make([]BuiltinMap, shardCount)

	for shardIndex := range shards {
		shards[shardIndex] = BuiltinMap{
			items: make(map[string]int),
		}
	}

	return &ShardedBuiltinMap{
		shards: shards,
	}
}

func (hashMap *ShardedBuiltinMap) getShard(key string) *BuiltinMap {
	shardIndex := int(hashString(key) % uint64(len(hashMap.shards)))
	return &hashMap.shards[shardIndex]
}

//func hashString(value string) uint64 {
//	hasher := fnv.New64a()
//	_, _ = hasher.Write([]byte(value))
//	return hasher.Sum64()
//}

func hashString(value string) uint64 {
	var hash uint64 = 14695981039346656037

	for index := 0; index < len(value); index++ {
		hash ^= uint64(value[index])
		hash *= 1099511628211
	}

	return hash
}
