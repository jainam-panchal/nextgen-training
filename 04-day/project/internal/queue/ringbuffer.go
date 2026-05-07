package queue

type RingBufferQueue[T any] struct {
	items []T
	head  int
	tail  int
	size  int
}

func NewRingBufferQueue[T any]() *RingBufferQueue[T] {
	return &RingBufferQueue[T]{
		items: make([]T, 16),
	}
}

func (q *RingBufferQueue[T]) Enqueue(value T) {
	if q.size == len(q.items) {
		q.grow()
	}

	q.items[q.tail] = value
	q.tail = (q.tail + 1) % len(q.items)
	q.size++
}

func (q *RingBufferQueue[T]) Dequeue() (T, bool) {
	var zero T

	if q.size == 0 {
		return zero, false
	}

	value := q.items[q.head]
	q.items[q.head] = zero

	q.head = (q.head + 1) % len(q.items)
	q.size--

	return value, true
}

func (q *RingBufferQueue[T]) Peek() (T, bool) {
	var zero T

	if q.size == 0 {
		return zero, false
	}

	return q.items[q.head], true
}

func (q *RingBufferQueue[T]) Len() int {
	return q.size
}

func (q *RingBufferQueue[T]) grow() {
	newItems := make([]T, len(q.items)*2)

	for i := 0; i < q.size; i++ {
		newItems[i] = q.items[(q.head+i)%len(q.items)]
	}

	q.items = newItems
	q.head = 0
	q.tail = q.size
}

var _ Queue[int] = (*RingBufferQueue[int])(nil)
