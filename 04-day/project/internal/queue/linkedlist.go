// Package queue
package queue

type node[T any] struct {
	value T
	next  *node[T]
}

type LinkedListQueue[T any] struct {
	head *node[T]
	tail *node[T]
	size int
}

func NewLinkedListQueue[T any]() *LinkedListQueue[T] {
	return &LinkedListQueue[T]{}
}

func (q *LinkedListQueue[T]) Enqueue(value T) {
	newNode := &node[T]{value: value}

	if q.size == 0 {
		q.head = newNode
		q.tail = newNode
	} else {
		q.tail.next = newNode
		q.tail = newNode
	}

	q.size++
}

func (q *LinkedListQueue[T]) Dequeue() (T, bool) {
	delNode := q.head

	if delNode == nil {
		var zero T
		return zero, false
	}

	q.head = delNode.next
	q.size--

	if q.size == 0 {
		q.tail = nil
	}

	return delNode.value, true
}

func (q *LinkedListQueue[T]) Peek() (T, bool) {
	if q.head == nil {
		var zero T
		return zero, false
	}

	return q.head.value, true
}

func (q *LinkedListQueue[T]) Len() int {
	return q.size
}
