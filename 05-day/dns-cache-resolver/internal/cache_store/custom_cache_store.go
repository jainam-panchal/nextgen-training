package cachestore

const (
	defaultBucketCount = 8
	maxLoadFactor      = 0.75
)

type entry[T any] struct {
	key   string
	value T
	next  *entry[T]
}

type CustomMapStore[T any] struct {
	buckets []*entry[T]
	size    int
}

func NewCustomMapStore[T any]() *CustomMapStore[T] {
	return &CustomMapStore[T]{
		buckets: make([]*entry[T], defaultBucketCount),
	}
}

// String to uint64
func hashString(s string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	var h uint64 = offset64
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

func (s *CustomMapStore[T]) bucketIndex(key string) int {
	return int(hashString(key) % uint64(len(s.buckets)))
}

func (s *CustomMapStore[T]) Get(key string) (T, bool) {
	idx := s.bucketIndex(key)

	for e := s.buckets[idx]; e != nil; e = e.next {
		if e.key == key {
			return e.value, true
		}
	}

	var zero T
	return zero, false
}

func (s *CustomMapStore[T]) resize(newBucketCount int) {
	oldBuckets := s.buckets

	s.buckets = make([]*entry[T], newBucketCount)
	s.size = 0

	for _, head := range oldBuckets {
		for e := head; e != nil; e = e.next {
			s.Set(e.key, e.value)
		}
	}
}

func (s *CustomMapStore[T]) Set(key string, value T) {
	if float64(s.size+1)/float64(len(s.buckets)) > maxLoadFactor {
		s.resize(len(s.buckets) * 2)
	}

	idx := s.bucketIndex(key)

	for e := s.buckets[idx]; e != nil; e = e.next {
		if e.key == key {
			e.value = value
			return
		}
	}

	s.buckets[idx] = &entry[T]{
		key:   key,
		value: value,
		next:  s.buckets[idx],
	}

	s.size++
}

func (s *CustomMapStore[T]) Delete(key string) bool {
	idx := s.bucketIndex(key)
	var prev *entry[T]

	for curr := s.buckets[idx]; curr != nil; curr = curr.next {
		if curr.key == key {
			if prev == nil {
				s.buckets[idx] = curr.next
			} else {
				prev.next = curr.next
			}
			s.size--
			return true
		}
		prev = curr
	}
	return false
}

func (s *CustomMapStore[T]) Len() int {
	return s.size
}

func (s *CustomMapStore[T]) Keys() []string {
	keys := make([]string, 0, s.size)

	for _, head := range s.buckets {
		for e := head; e != nil; e = e.next {
			keys = append(keys, e.key)
		}
	}

	return keys
}
