// Package container defines all the data structures to be used
package container

import (
	"fmt"
	"strings"
)

type Node[T any] struct {
	Data T
	Prev *Node[T]
	Next *Node[T]
}

type DoublyLinkedList[T any] struct {
	Head *Node[T]
	Tail *Node[T]

	Size int
}

func (dll *DoublyLinkedList[T]) Append(data T) *Node[T] {
	newNode := &Node[T]{
		Data: data,
	}

	if dll.Tail == nil {
		dll.Head = newNode
		dll.Tail = newNode
	} else {
		newNode.Prev = dll.Tail

		dll.Tail.Next = newNode
		dll.Tail = newNode
	}

	dll.Size++
	return newNode
}

func (dll *DoublyLinkedList[T]) RemoveAfter(node *Node[T]) {
	if node == nil {
		dll.Clear()
		return
	}

	delNodeCount := 0

	for tmp := node.Next; tmp != nil; tmp = tmp.Next {
		delNodeCount++
	}

	dll.Tail = node
	dll.Tail.Next = nil
	dll.Size -= delNodeCount
}

func (dll *DoublyLinkedList[T]) Clear() {
	dll.Head = nil
	dll.Tail = nil
	dll.Size = 0
}

func (dll *DoublyLinkedList[T]) Len() int {
	return dll.Size
}

func (dll *DoublyLinkedList[T]) String() string {
	if dll.Head == nil {
		return "[]"
	}

	var builder strings.Builder
	builder.WriteString("[")

	for node := dll.Head; node != nil; node = node.Next {
		builder.WriteString(fmt.Sprintf("%v", node.Data))

		if node.Next != nil {
			builder.WriteString(" <-> ")
		}

	}

	builder.WriteString("]")
	return builder.String()
}
