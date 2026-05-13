package main

type HashMap interface {
	Set(key string, value int)
	Get(key string) (int, bool)
	Delete(key string)
	Len() int
}

type BuiltinMap struct {
	items map[string]int
}

func NewBuiltinMap() *BuiltinMap {
	return &BuiltinMap{items: make(map[string]int)}
}

func (m *BuiltinMap) Set(key string, value int) {
	m.items[key] = value
}

func (m *BuiltinMap) Get(key string) (int, bool) {
	value, ok := m.items[key]
	return value, ok
}

func (m *BuiltinMap) Delete(key string) {
	delete(m.items, key)
}

func (m *BuiltinMap) Len() int {
	return len(m.items)
}

type ShardedBuiltinMap struct {
	shards []BuiltinMap
}

func NewShardedBuiltinMap(shardCount int) *ShardedBuiltinMap {
	if shardCount <= 0 {
		shardCount = 1
	}
	shards := make([]BuiltinMap, shardCount)
	for i := range shards {
		shards[i] = BuiltinMap{items: make(map[string]int)}
	}
	return &ShardedBuiltinMap{shards: shards}
}

func (m *ShardedBuiltinMap) Set(key string, value int) {
	shard := m.getShard(key)
	shard.Set(key, value)
}

func (m *ShardedBuiltinMap) Get(key string) (int, bool) {
	shard := m.getShard(key)
	return shard.Get(key)
}

func (m *ShardedBuiltinMap) Delete(key string) {
	shard := m.getShard(key)
	shard.Delete(key)
}

func (m *ShardedBuiltinMap) Len() int {
	total := 0
	for i := range m.shards {
		total += m.shards[i].Len()
	}
	return total
}

func (m *ShardedBuiltinMap) getShard(key string) *BuiltinMap {
	index := int(hashString(key) % uint64(len(m.shards)))
	return &m.shards[index]
}

func hashString(value string) uint64 {
	var hash uint64 = 14695981039346656037
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= 1099511628211
	}
	return hash
}

