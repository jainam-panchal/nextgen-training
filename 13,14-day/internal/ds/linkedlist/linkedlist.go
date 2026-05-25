package linkedlist

type Node[T any] struct {
	value T
	next  *Node[T]
}

type LinkedList[T any] struct {
	head *Node[T]
	tail *Node[T]
	size int
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{
		head: nil,
		tail: nil,
		size: 0,
	}
}

func (l *LinkedList[T]) AddRear(value T) {
	newNode := &Node[T]{
		value: value,
		next:  nil,
	}

	if l.tail == nil {
		l.head = newNode
		l.tail = newNode
		l.size++
		return
	}

	l.tail.next = newNode
	l.tail = newNode
	l.size++
}

func (l *LinkedList[T]) RemoveFront() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}

	value := l.head.value
	l.head = l.head.next
	l.size--

	if l.head == nil {
		l.tail = nil
	}

	return value, true
}

func (l *LinkedList[T]) RemoveRear() (T, bool) {
	if l.tail == nil {
		var zero T
		return zero, false
	}

	value := l.tail.value

	if l.head == l.tail {
		l.head = nil
		l.tail = nil
		l.size--
		return value, true
	}

	current := l.head
	for current.next != l.tail {
		current = current.next
	}

	current.next = nil
	l.tail = current
	l.size--

	return value, true
}

func (l *LinkedList[T]) PeekFront() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}

	return l.head.value, true
}

func (l *LinkedList[T]) PeekRear() (T, bool) {
	if l.tail == nil {
		var zero T
		return zero, false
	}

	return l.tail.value, true
}

func (l *LinkedList[T]) Size() int {
	return l.size
}

func (l *LinkedList[T]) IsEmpty() bool {
	return l.size == 0
}
