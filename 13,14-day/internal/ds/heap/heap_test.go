package heap

import "testing"

func TestHeap(t *testing.T) {
	heap := NewMaxHeap[int](func(a, b int) bool {
		return a < b
	})

	heap.Push(10)
	heap.Push(20)
	heap.Push(5)
	heap.Push(30)
	heap.Push(25)

	if val, ok := heap.Pop(); !ok || val != 30 {
		t.Fatal("Expected 30")
	}

	if val, ok := heap.Pop(); !ok || val != 25 {
		t.Fatal("Expected 30")
	}

	if val, ok := heap.Pop(); !ok || val != 20 {
		t.Fatal("Expected 30")
	}

	if val, ok := heap.Pop(); !ok || val != 10 {
		t.Fatal("Expected 30")
	}

	if val, ok := heap.Pop(); !ok || val != 5 {
		t.Fatal("Expected 30")
	}

	if heap.Size() != 0 {
		t.Fatal("Expected empty heap")
	}

	if !heap.IsEmpty() {
		t.Fatal("Expected empty heap from func")
	}

}
