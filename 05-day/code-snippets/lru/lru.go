// Package lru
package lru

type LRU[T any] interface {
	Put(string, T)
	Get(string) (T, bool)
	Len() int
}

type Node[T any] struct {
	key   string
	value T

	prev *Node[T]
	next *Node[T]
}

type LRUCache[T any] struct {
	capacity int
	size     int

	items map[string]*Node[T]

	head *Node[T]
	tail *Node[T]
}

func NewLRUCache[T any](cap int) *LRUCache[T] {
	return &LRUCache[T]{
		capacity: cap,
		items:    make(map[string]*Node[T]),
	}
}

func (L *LRUCache[T]) Put(s string, t T) {
	if s == "" || L.capacity <= 0 {
		return
	}

	if existingNode, ok := L.items[s]; ok {
		existingNode.value = t // Update value
		L.detach(existingNode)
		L.attachFirstNode(existingNode)
		return
	}

	if L.size >= L.capacity {
		L.RemoveLastNode()
	}

	newNode := &Node[T]{key: s, value: t}
	L.attachFirstNode(newNode)
}

func (L *LRUCache[T]) Get(s string) (T, bool) {
	node, ok := L.items[s]
	if !ok {
		var zero T
		return zero, false
	}

	L.detach(node)
	L.attachFirstNode(node)

	return node.value, true
}

func (L *LRUCache[T]) Len() int {
	return L.size
}

func (L *LRUCache[T]) RemoveLastNode() {
	if L.tail != nil {
		L.detach(L.tail)
	}
}

func (L *LRUCache[T]) attachFirstNode(node *Node[T]) {
	node.prev = nil
	node.next = L.head

	if L.head == nil {
		L.head = node
		L.tail = node
	} else {
		L.head.prev = node
		L.head = node
	}

	L.items[node.key] = node
	L.size++
}

func (L *LRUCache[T]) detach(node *Node[T]) {

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		L.head = node.next // Node was head
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		L.tail = node.prev // Node was tail
	}

	node.next = nil
	node.prev = nil

	delete(L.items, node.key)
	L.size--
}
