package ds

// RingBuffer is a generic fixed-capacity ring (circular) buffer.
// When full, new Appends overwrite the oldest element.
// Used by the engine to track rolling congestion history per road.
type RingBuffer[T any] struct {
	buf   []T
	start int // index of oldest element
	size  int // number of elements currently stored
	cap   int // maximum capacity
}

// NewRingBuffer creates a ring buffer with the given capacity (must be > 0).
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		panic("capacity must be > 0")
	}
	return &RingBuffer[T]{buf: make([]T, capacity), cap: capacity}
}

// Append adds an element. If the buffer is full, the oldest element is overwritten.
func (r *RingBuffer[T]) Append(v T) {
	if r.size < r.cap {
		r.buf[(r.start+r.size)%r.cap] = v
		r.size++
		return
	}
	r.buf[r.start] = v
	r.start = (r.start + 1) % r.cap
}

// Items returns all elements in FIFO order (oldest first).
func (r *RingBuffer[T]) Items() []T {
	out := make([]T, r.size)
	for i := 0; i < r.size; i++ {
		out[i] = r.buf[(r.start+i)%r.cap]
	}
	return out
}

// Len returns the number of elements currently stored.
func (r *RingBuffer[T]) Len() int { return r.size }

// Capacity returns the maximum capacity.
func (r *RingBuffer[T]) Capacity() int { return r.cap }
