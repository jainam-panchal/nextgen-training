package queue

import (
	"errors"
)

var (
	ErrEmptyQueue = errors.New("queue is empty")
	ErrFullQueue  = errors.New("queue is full")
)

type CircularQueue[T any] struct {
	data []T
	head int
	tail int
	size int
	cap  int
}

func NewCircularQueue[T any](cap int) *CircularQueue[T] {
	return &CircularQueue[T]{
		data: make([]T, 0, cap),
		head: 0,
		tail: 0,
		size: 0,
		cap:  cap,
	}
}

func (c *CircularQueue[T]) Enqueue(t T) error {
	if c.size == c.cap {
		return ErrFullQueue
	}

	c.data[c.tail] = t
	c.tail = (c.tail + 1) % c.cap
	c.size++
	return nil
}

func (c *CircularQueue[T]) EnqueueOverwrite(t T) {
	c.data[c.tail] = t
	c.tail = (c.tail + 1) % c.cap

	if c.size == c.cap {
		c.head = (c.head + 1) % c.cap
	} else {
		c.size++
	}
}

func (c *CircularQueue[T]) Dequeue() (T, error) {
	if c.size == 0 {
		var zero T
		return zero, ErrEmptyQueue
	}

	v := c.data[c.head]
	var zero T
	c.data[c.head] = zero
	c.head = (c.head + 1) % c.cap
	c.size--
	return v, nil

}

func (c *CircularQueue[T]) Peek() (T, error) {
	if c.size == 0 {
		var zero T
		return zero, ErrEmptyQueue
	}

	return c.data[c.head], nil
}

func (c *CircularQueue[T]) IsFull() bool {
	return c.size == c.cap
}

func (c *CircularQueue[T]) IsEmpty() bool {
	return c.size == 0
}

func (c *CircularQueue[T]) Size() int {
	return c.size
}

func (c *CircularQueue[T]) Cap() int {
	return c.cap
}

func (c *CircularQueue[T]) Len() int {
	return c.size
}
