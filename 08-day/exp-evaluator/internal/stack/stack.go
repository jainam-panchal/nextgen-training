package stack

type Stack[T any] struct {
	items []T
}

func NewStack[T any](len int) *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0, len),
	}
}

func (s *Stack[T]) Push(val T) {
	s.items = append(s.items, val)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T

	if len(s.items) == 0 {
		return zero, false
	}

	lastIdx := len(s.items) - 1
	data := s.items[lastIdx]

	s.items[lastIdx] = zero
	s.items = s.items[:lastIdx]

	return data, true
}

func (s *Stack[T]) Peek() (T, bool) {
	var zero T

	if len(s.items) == 0 {
		return zero, false
	}

	lastIdx := len(s.items) - 1
	data := s.items[lastIdx]
	return data, true
}

func (s *Stack[T]) MustPeek() T {
	if len(s.items) == 0 {
		panic("empty stack")
	}

	return s.items[len(s.items)-1]
}

func (s *Stack[T]) MustPop() T {
	if len(s.items) == 0 {
		panic("empty stack")
	}

	var zero T
	lastIdx := len(s.items) - 1
	data := s.items[lastIdx]
	s.items[lastIdx] = zero
	s.items = s.items[:lastIdx]

	return data
}

func (s *Stack[T]) Size() int {
	return len(s.items)
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}
