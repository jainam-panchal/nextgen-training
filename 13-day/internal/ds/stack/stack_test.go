package stack

import "testing"

func TestStack(t *testing.T) {
	stack := NewStack[int]()

	stack.Push(10)
	stack.Push(20)
	stack.Push(30)

	if val, ok := stack.Pop(); !ok || val != 30 {
		t.Fatal("Expected 30")
	}

	if val, ok := stack.Pop(); !ok || val != 20 {
		t.Fatal("Expected 20")
	}

	if val, ok := stack.Pop(); !ok || val != 10 {
		t.Fatal("Expected 10")
	}

	if stack.Size() != 0 {
		t.Fatal("Expected empty stack")
	}

	if !stack.IsEmpty() {
		t.Fatal("Expected empty stack from func")
	}

	if _, ok := stack.Peek(); ok {
		t.Fatal("Expected empty stack")
	}

}
