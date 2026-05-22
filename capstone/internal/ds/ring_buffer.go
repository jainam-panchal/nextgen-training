package ds

type RingBuffer[T any] struct {
	buf   []T
	start int
	size  int
	cap   int
}

func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		panic("capacity must be > 0")
	}
	return &RingBuffer[T]{buf: make([]T, capacity), cap: capacity}
}

func (r *RingBuffer[T]) Append(v T) {
	if r.size < r.cap {
		r.buf[(r.start+r.size)%r.cap] = v
		r.size++
		return
	}
	r.buf[r.start] = v
	r.start = (r.start + 1) % r.cap
}

func (r *RingBuffer[T]) Items() []T {
	out := make([]T, r.size)
	for i := 0; i < r.size; i++ {
		out[i] = r.buf[(r.start+i)%r.cap]
	}
	return out
}

func (r *RingBuffer[T]) Len() int { return r.size }

func (r *RingBuffer[T]) Capacity() int { return r.cap }
