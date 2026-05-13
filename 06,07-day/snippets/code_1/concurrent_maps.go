package main

import "sync"

type ConcurrentMap interface {
	Set(key string, value int)
	Get(key string) (int, bool)
	Delete(key string)
}

// LockedBuiltinMap uses a single RWMutex over one Go map.
type LockedBuiltinMap struct {
	mu    sync.RWMutex
	items map[string]int
}

func NewLockedBuiltinMap() *LockedBuiltinMap {
	return &LockedBuiltinMap{
		items: make(map[string]int),
	}
}

func (m *LockedBuiltinMap) Set(key string, value int) {
	m.mu.Lock()
	m.items[key] = value
	m.mu.Unlock()
}

func (m *LockedBuiltinMap) Get(key string) (int, bool) {
	m.mu.RLock()
	value, ok := m.items[key]
	m.mu.RUnlock()
	return value, ok
}

func (m *LockedBuiltinMap) Delete(key string) {
	m.mu.Lock()
	delete(m.items, key)
	m.mu.Unlock()
}

type shardWithLock struct {
	mu    sync.RWMutex
	items map[string]int
}

// LockedShardedBuiltinMap uses per-shard RWMutex for lower lock contention.
type LockedShardedBuiltinMap struct {
	shards []shardWithLock
}

func NewLockedShardedBuiltinMap(shardCount int) *LockedShardedBuiltinMap {
	shards := make([]shardWithLock, shardCount)
	for i := range shards {
		shards[i] = shardWithLock{
			items: make(map[string]int),
		}
	}

	return &LockedShardedBuiltinMap{
		shards: shards,
	}
}

func (m *LockedShardedBuiltinMap) getShard(key string) *shardWithLock {
	shardIndex := int(hashString(key) % uint64(len(m.shards)))
	return &m.shards[shardIndex]
}

func (m *LockedShardedBuiltinMap) Set(key string, value int) {
	shard := m.getShard(key)
	shard.mu.Lock()
	shard.items[key] = value
	shard.mu.Unlock()
}

func (m *LockedShardedBuiltinMap) Get(key string) (int, bool) {
	shard := m.getShard(key)
	shard.mu.RLock()
	value, ok := shard.items[key]
	shard.mu.RUnlock()
	return value, ok
}

func (m *LockedShardedBuiltinMap) Delete(key string) {
	shard := m.getShard(key)
	shard.mu.Lock()
	delete(shard.items, key)
	shard.mu.Unlock()
}

// SyncMapAdapter wraps sync.Map behind the same benchmark interface.
type SyncMapAdapter struct {
	items sync.Map
}

func NewSyncMapAdapter() *SyncMapAdapter {
	return &SyncMapAdapter{}
}

func (m *SyncMapAdapter) Set(key string, value int) {
	m.items.Store(key, value)
}

func (m *SyncMapAdapter) Get(key string) (int, bool) {
	value, ok := m.items.Load(key)
	if !ok {
		return 0, false
	}

	intValue, ok := value.(int)
	if !ok {
		return 0, false
	}

	return intValue, true
}

func (m *SyncMapAdapter) Delete(key string) {
	m.items.Delete(key)
}
