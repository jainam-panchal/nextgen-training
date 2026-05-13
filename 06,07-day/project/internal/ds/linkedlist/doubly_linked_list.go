package linkedlist

import (
	"sync"

	appErrors "ride-sharing/internal/errors"
)

type Node[T any] struct {
	value T
	prev  *Node[T]
	next  *Node[T]
}

func (n *Node[T]) Value() T {
	return n.value
}

type DoublyLinkedList[T any] struct {
	mu     sync.RWMutex
	head   *Node[T]
	tail   *Node[T]
	length int
}

func NewDoublyLinkedList[T any]() *DoublyLinkedList[T] {
	return &DoublyLinkedList[T]{}
}

func (l *DoublyLinkedList[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.length
}

func (l *DoublyLinkedList[T]) IsEmpty() bool {
	return l.Len() == 0
}

func (l *DoublyLinkedList[T]) PushBack(value T) *Node[T] {
	l.mu.Lock()
	defer l.mu.Unlock()

	newNode := &Node[T]{
		value: value,
	}

	if l.length == 0 {
		l.head = newNode
		l.tail = newNode
		l.length++
		return newNode
	}

	newNode.prev = l.tail
	l.tail.next = newNode
	l.tail = newNode
	l.length++

	return newNode
}

func (l *DoublyLinkedList[T]) Remove(node *Node[T]) (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var zero T

	if node == nil {
		return zero, appErrors.ErrInvalidLinkedListNode
	}

	if l.length == 0 {
		return zero, appErrors.ErrEmptyLinkedList
	}

	if !l.containsNode(node) {
		return zero, appErrors.ErrInvalidLinkedListNode
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}

	l.length--

	removedValue := node.value

	node.prev = nil
	node.next = nil

	return removedValue, nil
}

func (l *DoublyLinkedList[T]) containsNode(target *Node[T]) bool {
	for currentNode := l.head; currentNode != nil; currentNode = currentNode.next {
		if currentNode == target {
			return true
		}
	}

	return false
}

func (l *DoublyLinkedList[T]) Values() []T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	values := make([]T, 0, l.length)

	for currentNode := l.head; currentNode != nil; currentNode = currentNode.next {
		values = append(values, currentNode.value)
	}

	return values
}
