package heap

type MaxHeap[T any] struct {
	items []T
	less  func(a, b T) bool // less returns true if a < b
}

func NewMaxHeap[T any](less func(a, b T) bool) *MaxHeap[T] {
	return &MaxHeap[T]{
		items: make([]T, 0),
		less:  less,
	}
}

func (h *MaxHeap[T]) Size() int {
	return len(h.items)
}

func (h *MaxHeap[T]) IsEmpty() bool {
	return len(h.items) == 0
}

func (h *MaxHeap[T]) Push(item T) {
	h.items = append(h.items, item)
	h.heapifyUp(len(h.items) - 1)
}

func (h *MaxHeap[T]) heapifyUp(i int) {
	for i > 0 {
		parent := h.parent(i)
		if h.less(h.items[parent], h.items[i]) {
			h.items[i], h.items[parent] = h.items[parent], h.items[i]
			i = parent
		} else {
			break
		}
	}
}

func (h *MaxHeap[T]) Pop() (T, bool) {
	if h.IsEmpty() {
		var zero T
		return zero, false
	}

	root := h.items[0]
	h.items[0] = h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]

	if len(h.items) > 0 {
		h.heapifyDown(0)
	}

	return root, true
}

func (h *MaxHeap[T]) heapifyDown(i int) {
	for i < len(h.items) {
		left := h.left(i)
		right := h.right(i)
		toCompare := i

		if left < len(h.items) && h.less(h.items[toCompare], h.items[left]) {
			toCompare = left
		}
		if right < len(h.items) && h.less(h.items[toCompare], h.items[right]) {
			toCompare = right
		}
		if toCompare == i {
			break
		}

		h.items[i], h.items[toCompare] = h.items[toCompare], h.items[i]
		i = toCompare
	}
}

func (h *MaxHeap[T]) Peek() (T, bool) {
	if h.IsEmpty() {
		var zero T
		return zero, false
	}
	return h.items[0], true
}

func (h *MaxHeap[T]) parent(i int) int {
	return (i - 1) / 2
}
func (h *MaxHeap[T]) left(i int) int {
	return 2*i + 1
}
func (h *MaxHeap[T]) right(i int) int {
	return 2*i + 2
}
