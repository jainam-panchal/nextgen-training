package linkedlist

import (
	stderrors "errors"
	"testing"

	appErrors "ride-sharing/internal/errors"
)

// Tests push order, remove behavior, and values projection.
func TestDoublyLinkedListBasicFlow(t *testing.T) {
	list := NewDoublyLinkedList[int]()

	n1 := list.PushBack(10)
	n2 := list.PushBack(20)
	n3 := list.PushBack(30)
	_ = n3

	if list.Len() != 3 {
		t.Fatalf("expected len 3, got %d", list.Len())
	}

	removed, err := list.Remove(n2)
	if err != nil {
		t.Fatalf("remove middle failed: %v", err)
	}
	if removed != 20 {
		t.Fatalf("expected removed value 20, got %d", removed)
	}

	values := list.Values()
	if len(values) != 2 || values[0] != 10 || values[1] != 30 {
		t.Fatalf("expected values [10 30], got %v", values)
	}

	_, err = list.Remove(n1)
	if err != nil {
		t.Fatalf("remove head failed: %v", err)
	}
	if list.Len() != 1 {
		t.Fatalf("expected len 1 after head remove, got %d", list.Len())
	}
}

// Tests invalid remove cases (nil node, stale node, empty list).
func TestDoublyLinkedListRemoveErrors(t *testing.T) {
	list := NewDoublyLinkedList[int]()

	if _, err := list.Remove(nil); !stderrors.Is(err, appErrors.ErrInvalidLinkedListNode) {
		t.Fatalf("expected ErrInvalidLinkedListNode for nil, got %v", err)
	}

	other := NewDoublyLinkedList[int]()
	foreignNode := other.PushBack(1)
	if _, err := list.Remove(foreignNode); !stderrors.Is(err, appErrors.ErrEmptyLinkedList) {
		t.Fatalf("expected ErrEmptyLinkedList when list is empty, got %v", err)
	}

	n := list.PushBack(10)
	if _, err := list.Remove(n); err != nil {
		t.Fatalf("remove existing node failed: %v", err)
	}
	if _, err := list.Remove(n); !stderrors.Is(err, appErrors.ErrEmptyLinkedList) {
		t.Fatalf("expected ErrEmptyLinkedList after removing last node, got %v", err)
	}
}
