package ds

import "container/heap"

type Item[T any] struct {
	Value    T
	Priority float64
	index    int
}

type PriorityQueue[T any] struct {
	items []*Item[T]
}

func NewPriorityQueue[T any]() *PriorityQueue[T] {
	pq := &PriorityQueue[T]{items: make([]*Item[T], 0)}
	return pq
}

func (pq *PriorityQueue[T]) Len() int { return len(pq.items) }

func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.items[i].Priority < pq.items[j].Priority
}

func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *PriorityQueue[T]) Push(x any) {
	item := x.(*Item[T])
	item.index = len(pq.items)
	pq.items = append(pq.items, item)
}

func (pq *PriorityQueue[T]) Pop() any {
	old := pq.items
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[n-1]
	pq.items = old[:n-1]
	item.index = -1
	return item
}

func (pq *PriorityQueue[T]) PushItem(value T, priority float64) *Item[T] {
	item := &Item[T]{Value: value, Priority: priority}
	heap.Push(pq, item)
	return item
}

func (pq *PriorityQueue[T]) PopItem() (value T, priority float64, ok bool) {
	if pq.Len() == 0 {
		var zero T
		return zero, 0, false
	}
	item := heap.Pop(pq).(*Item[T])
	return item.Value, item.Priority, true
}

func (pq *PriorityQueue[T]) Peek() (value T, priority float64, ok bool) {
	if pq.Len() == 0 {
		var zero T
		return zero, 0, false
	}
	it := pq.items[0]
	return it.Value, it.Priority, true
}

func (pq *PriorityQueue[T]) Update(item *Item[T], priority float64) {
	item.Priority = priority
	heap.Fix(pq, item.index)
}

func (pq *PriorityQueue[T]) Init() {
	heap.Init(pq)
}
