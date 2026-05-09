package cachestore

type InbuiltMapStore[T any] struct {
	mp map[string]T
}

func NewInbuiltMapStore[T any]() *InbuiltMapStore[T] {
	return &InbuiltMapStore[T]{
		mp: make(map[string]T),
	}
}

func (mp *InbuiltMapStore[T]) Get(key string) (T, bool) {
	val, ok := mp.mp[key]
	return val, ok
}

func (mp *InbuiltMapStore[T]) Set(key string, val T) {
	mp.mp[key] = val
}

func (mp *InbuiltMapStore[T]) Delete(key string) bool {
	if _, ok := mp.mp[key]; ok {
		delete(mp.mp, key)
		return true
	}
	return false
}

func (mp *InbuiltMapStore[T]) Len() int {
	return len(mp.mp)
}

func (mp *InbuiltMapStore[T]) Keys() []string {
	keys := make([]string, 0, len(mp.mp))
	for k := range mp.mp {
		keys = append(keys, k)
	}
	return keys
}
