package main

import "fmt"

type CircularBufferList[T any] struct {
	data []T
	size int
	head int
	tail int
}

func NewCircularBufferList[T any](cap int) *CircularBufferList[T] {
	if cap <= 0 {
		panic("cap must be greater than zero")
	}

	return &CircularBufferList[T]{
		data: make([]T, cap),
	}
}

func (b *CircularBufferList[T]) Insert(value T) {
	b.data[b.tail] = value

	if b.isFull() {
		b.head = (b.head + 1) % len(b.data)
	} else {
		b.size++
	}

	b.tail = (b.tail + 1) % len(b.data)
}

func (b *CircularBufferList[T]) Pop() (T, bool) {
	var zero T

	if b.isEmpty() {
		return zero, false
	}

	value := b.data[b.head]
	b.head = (b.head + 1) % len(b.data)
	b.size--

	return value, true
}

func (b *CircularBufferList[T]) Peek() (T, bool) {
	var zero T

	if b.size == 0 {
		return zero, false
	}

	return b.data[b.head], true
}

func (b *CircularBufferList[T]) Values() []T {
	results := make([]T, 0, b.size)

	for i := 0; i < b.size; i++ {
		idx := (b.head + i) % len(b.data)
		results = append(results, b.data[idx])
	}

	return results
}

func (b *CircularBufferList[T]) isEmpty() bool {
	return b.size == 0
}

func (b *CircularBufferList[T]) isFull() bool {
	return b.size == len(b.data)
}

func (b *CircularBufferList[T]) Capacity() int {
	return len(b.data)
}

func (b *CircularBufferList[T]) Len() int {
	return b.size
}

func main() {
	buf := NewCircularBufferList[int](3)

	buf.Insert(10)
	buf.Insert(20)
	buf.Insert(30)

	fmt.Println(buf.Values()) // [10 20 30]

	buf.Insert(40)

	fmt.Println(buf.Values()) // [20 30 40]

	v, ok := buf.Pop()
	fmt.Println(v, ok)        // 20 true
	fmt.Println(buf.Values()) // [30 40]

	buf.Insert(50)
	buf.Insert(60)

	fmt.Println(buf.Values()) // [40 50 60]
}
