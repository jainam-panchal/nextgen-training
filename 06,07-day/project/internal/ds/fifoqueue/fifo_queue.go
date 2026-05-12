package fifoqueue

import appErrors "ride-sharing/internal/errors"

type Queue[T any] struct {
	items []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items: make([]T, 0),
	}
}

func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

func (q *Queue[T]) Dequeue() (T, error) {
	var zero T

	if len(q.items) == 0 {
		return zero, appErrors.ErrEmptyQueue
	}

	frontItem := q.items[0]

	q.items[0] = zero
	q.items = q.items[1:]

	return frontItem, nil
}

func (q *Queue[T]) Peek() (T, error) {
	var zero T

	if len(q.items) == 0 {
		return zero, appErrors.ErrEmptyQueue
	}

	return q.items[0], nil
}

func (q *Queue[T]) Len() int {
	return len(q.items)
}

func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *Queue[T]) Values() []T {
	values := make([]T, len(q.items))
	copy(values, q.items)

	return values
}
