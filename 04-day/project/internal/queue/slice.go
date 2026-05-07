package queue

type SliceQueue[T any] struct {
	items []T
}

func NewSliceQueue[T any]() *SliceQueue[T] {
	return &SliceQueue[T]{}
}

func (q *SliceQueue[T]) Enqueue(value T) {
	q.items = append(q.items, value)
}

func (q *SliceQueue[T]) Dequeue() (T, bool) {
	var zero T

	if len(q.items) == 0 {
		return zero, false
	}

	value := q.items[0]

	var empty T
	q.items[0] = empty
	q.items = q.items[1:]

	return value, true
}

func (q *SliceQueue[T]) Peek() (T, bool) {
	var zero T

	if len(q.items) == 0 {
		return zero, false
	}

	return q.items[0], true
}

func (q *SliceQueue[T]) Len() int {
	return len(q.items)
}

var _ Queue[int] = (*SliceQueue[int])(nil)