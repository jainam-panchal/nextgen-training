package heap

import appErrors "ride-sharing/internal/errors"

type LessFunc[T any] func(first T, second T) bool

type MinHeap[T any] struct {
	items []T
	less  LessFunc[T]
}

func NewMinHeap[T any](less LessFunc[T]) *MinHeap[T] {
	return &MinHeap[T]{
		items: make([]T, 0),
		less:  less,
	}
}

func (h *MinHeap[T]) Push(item T) {
	h.items = append(h.items, item)
	h.heapifyUp(len(h.items) - 1)
}

func (h *MinHeap[T]) Pop() (T, error) {
	var zero T

	if len(h.items) == 0 {
		return zero, appErrors.ErrEmptyQueue
	}

	rootItem := h.items[0]
	lastIndex := len(h.items) - 1

	h.items[0] = h.items[lastIndex]
	h.items = h.items[:lastIndex]

	if len(h.items) > 0 {
		h.heapifyDown(0)
	}

	return rootItem, nil
}

func (h *MinHeap[T]) Peek() (T, error) {
	var zero T

	if len(h.items) == 0 {
		return zero, appErrors.ErrEmptyQueue
	}

	return h.items[0], nil
}

func (h *MinHeap[T]) Len() int {
	return len(h.items)
}

func (h *MinHeap[T]) IsEmpty() bool {
	return len(h.items) == 0
}

func (h *MinHeap[T]) heapifyUp(index int) {
	for index > 0 {
		parentIndex := getParentIndex(index)

		if !h.less(h.items[index], h.items[parentIndex]) {
			break
		}

		h.swap(index, parentIndex)
		index = parentIndex
	}
}

func (h *MinHeap[T]) heapifyDown(index int) {
	lastIndex := len(h.items) - 1

	for {
		leftIndex := getLeftChildIndex(index)
		rightIndex := getRightChildIndex(index)
		highestPriorityIndex := index

		if leftIndex <= lastIndex &&
			h.less(h.items[leftIndex], h.items[highestPriorityIndex]) {
			highestPriorityIndex = leftIndex
		}

		if rightIndex <= lastIndex &&
			h.less(h.items[rightIndex], h.items[highestPriorityIndex]) {
			highestPriorityIndex = rightIndex
		}

		if highestPriorityIndex == index {
			break
		}

		h.swap(index, highestPriorityIndex)
		index = highestPriorityIndex
	}
}

func (h *MinHeap[T]) swap(firstIndex int, secondIndex int) {
	h.items[firstIndex], h.items[secondIndex] = h.items[secondIndex], h.items[firstIndex]
}

func getParentIndex(childIndex int) int {
	return (childIndex - 1) / 2
}

func getLeftChildIndex(parentIndex int) int {
	return parentIndex*2 + 1
}

func getRightChildIndex(parentIndex int) int {
	return parentIndex*2 + 2
}
