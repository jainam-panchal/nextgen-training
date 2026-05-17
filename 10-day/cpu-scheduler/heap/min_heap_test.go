package heap

import (
	"errors"
	"testing"
)

// Tests push/pop/peek behavior for min-heap ordering.
func TestMinHeapOrder(t *testing.T) {
	h := NewMinHeap(func(a, b int) bool { return a < b })

	h.Push(5)
	h.Push(1)
	h.Push(3)

	if !h.Verify() {
		t.Fatal("heap property violated after push")
	}

	top, err := h.Peek()
	if err != nil {
		t.Fatalf("peek failed: %v", err)
	}

	if top != 1 {
		t.Fatalf("expected top 1, got %d", top)
	}

	first, err := h.Pop()
	if err != nil {
		t.Fatalf("first pop failed: %v", err)
	}

	second, err := h.Pop()
	if err != nil {
		t.Fatalf("second pop failed: %v", err)
	}

	third, err := h.Pop()
	if err != nil {
		t.Fatalf("third pop failed: %v", err)
	}

	if first != 1 || second != 3 || third != 5 {
		t.Fatalf("expected pop order 1,3,5 got %d,%d,%d", first, second, third)
	}
}

// Tests empty-heap error behavior.
func TestMinHeapEmptyErrors(t *testing.T) {
	h := NewMinHeap(func(a, b int) bool { return a < b })

	if _, err := h.Peek(); !errors.Is(err, ErrEmptyHeap) {
		t.Fatalf("expected ErrEmptyHeap from peek, got %v", err)
	}

	if _, err := h.Pop(); !errors.Is(err, ErrEmptyHeap) {
		t.Fatalf("expected ErrEmptyHeap from pop, got %v", err)
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

	top, err := h.Peek()
	if err != nil {
		t.Fatalf("peek failed: %v", err)
	}

	if top != 1 {
		t.Fatalf("heap mutated through Items copy, expected top 1 got %d", top)
	}
}

// Tests heap property after each mutation.
func TestMinHeapVerifyAfterOperations(t *testing.T) {
	h := NewMinHeap(func(a, b int) bool { return a < b })

	values := []int{9, 4, 7, 1, 3, 6, 2}

	for _, value := range values {
		h.Push(value)

		if !h.Verify() {
			t.Fatalf("heap property violated after pushing %d", value)
		}
	}

	for !h.IsEmpty() {
		_, err := h.Pop()
		if err != nil {
			t.Fatalf("pop failed: %v", err)
		}

		if !h.Verify() {
			t.Fatal("heap property violated after pop")
		}
	}
}
