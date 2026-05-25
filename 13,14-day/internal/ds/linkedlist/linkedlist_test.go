package linkedlist

import "testing"

func TestLinkedList(t *testing.T) {
	ll := NewLinkedList[int]()

	ll.AddRear(10)
	ll.AddRear(20)
	ll.AddRear(30)

	if val, ok := ll.PeekFront(); !ok || val != 10 {
		t.Errorf("Expected front value 10, got %d", val)
	}

	if val, ok := ll.PeekRear(); !ok || val != 30 {
		t.Errorf("Expected rear value 30, got %d", val)
	}

	if val, ok := ll.RemoveFront(); !ok || val != 10 {
		t.Errorf("Expected removed front value 10, got %d", val)
	}

	if val, ok := ll.RemoveRear(); !ok || val != 30 {
		t.Errorf("Expected removed rear value 30, got %d", val)
	}

	if val, ok := ll.RemoveFront(); !ok || val != 20 {
		t.Errorf("Expected removed front value 20, got %d", val)
	}

	if _, ok := ll.RemoveFront(); ok {
		t.Error("Expected empty list")
	}

	if ll.Size() != 0 {
		t.Errorf("Expected size 0, got %d", ll.Size())
	}

	if !ll.IsEmpty() {
		t.Error("Expected empty list")
	}
}
