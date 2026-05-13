package heap

import (
	stderrors "errors"
	"testing"

	appErrors "ride-sharing/internal/errors"
)

// Tests push/pop/peek behavior for min-heap ordering.
func TestMinHeapOrder(t *testing.T) {
	h := NewMinHeap(func(a, b int) bool { return a < b })

	h.Push(5)
	h.Push(1)
	h.Push(3)

	top, err := h.Peek()
	if err != nil {
		t.Fatalf("peek failed: %v", err)
	}
	if top != 1 {
		t.Fatalf("expected top 1, got %d", top)
	}

	first, _ := h.Pop()
	second, _ := h.Pop()
	third, _ := h.Pop()
	if first != 1 || second != 3 || third != 5 {
		t.Fatalf("expected pop order 1,3,5 got %d,%d,%d", first, second, third)
	}
}

// Tests empty-heap error behavior.
func TestMinHeapEmptyErrors(t *testing.T) {
	h := NewMinHeap(func(a, b int) bool { return a < b })

	if _, err := h.Peek(); !stderrors.Is(err, appErrors.ErrEmptyQueue) {
		t.Fatalf("expected ErrEmptyQueue from peek, got %v", err)
	}
	if _, err := h.Pop(); !stderrors.Is(err, appErrors.ErrEmptyQueue) {
		t.Fatalf("expected ErrEmptyQueue from pop, got %v", err)
	}
}

// Tests Items returns a copy, not backing slice alias.
func TestMinHeapItemsCopy(t *testing.T) {
	h := NewMinHeap(func(a, b int) bool { return a < b })
	h.Push(2)
	h.Push(1)

	items := h.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	items[0] = 999

	top, _ := h.Peek()
	if top != 1 {
		t.Fatalf("heap mutated through Items copy, expected top 1 got %d", top)
	}
}
