// Package cache -> implements LRU using container/list
package cache

import "container/list"

type LruCache[T any] struct {
	cap   int
	evict *list.List
	mp    map[string]*list.Element
}

type entry[T any] struct {
	key   string
	value T
}

func NewLruCache[T any](cap int) *LruCache[T] {
	return &LruCache[T]{
		cap:   cap,
		evict: list.New(),
		mp:    make(map[string]*list.Element),
	}
}

func (l *LruCache[T]) Get(key string) (T, bool) {
	if element, ok := l.mp[key]; ok {
		l.evict.MoveToFront(element)
		return element.Value.(entry[T]).value, true
	}

	var zero T
	return zero, false
}

func (l *LruCache[T]) Put(key string, value T) {
	if l.cap <= 0 {
		return
	}

	if element, ok := l.mp[key]; ok {
		l.evict.MoveToFront(element)
		element.Value.(*entry[T]).value = value
		return
	}

	ent := &entry[T]{key, value}
	element := l.evict.PushFront(ent)
	l.mp[key] = element

	if l.evict.Len() > l.cap {
		l.removeOldest()
	}
}

func (l *LruCache[T]) removeOldest() {
	element := l.evict.Back()
	if element != nil {
		l.evict.Remove(element)
		kv := element.Value.(*entry[T]).key
		delete(l.mp, kv)
	}
}

func (l *LruCache[T]) Len() int {
	return l.evict.Len()
}
