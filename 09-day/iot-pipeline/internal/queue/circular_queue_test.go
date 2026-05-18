package queue

import (
	"errors"
	"testing"
)

func TestCircularQueue_EmptyQueue(t *testing.T) {
	q := &CircularQueue[int]{
		data: make([]int, 3),
		cap:  3,
	}

	if !q.IsEmpty() {
		t.Fatal("expected queue to be empty")
	}

	if q.IsFull() {
		t.Fatal("expected queue to not be full")
	}

	if q.Size() != 0 {
		t.Fatalf("Size() = %d, want 0", q.Size())
	}

	if q.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", q.Len())
	}

	if q.Cap() != 3 {
		t.Fatalf("Cap() = %d, want 3", q.Cap())
	}
}

func TestCircularQueue_EnqueueDequeue(t *testing.T) {
	q := &CircularQueue[int]{
		data: make([]int, 3),
		cap:  3,
	}

	if err := q.Enqueue(10); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	if err := q.Enqueue(20); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	if q.Size() != 2 {
		t.Fatalf("Size() = %d, want 2", q.Size())
	}

	first, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}
	if first != 10 {
		t.Fatalf("Dequeue() = %d, want 10", first)
	}

	second, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}
	if second != 20 {
		t.Fatalf("Dequeue() = %d, want 20", second)
	}

	if !q.IsEmpty() {
		t.Fatal("expected queue to be empty after dequeuing all items")
	}
}

func TestCircularQueue_Peek(t *testing.T) {
	q := &CircularQueue[string]{
		data: make([]string, 2),
		cap:  2,
	}

	if err := q.Enqueue("alpha"); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	got, err := q.Peek()
	if err != nil {
		t.Fatalf("Peek() error = %v", err)
	}
	if got != "alpha" {
		t.Fatalf("Peek() = %q, want %q", got, "alpha")
	}

	if q.Size() != 1 {
		t.Fatalf("Peek() should not remove item; Size() = %d, want 1", q.Size())
	}
}

func TestCircularQueue_DequeueEmpty(t *testing.T) {
	q := &CircularQueue[int]{
		data: make([]int, 2),
		cap:  2,
	}

	_, err := q.Dequeue()
	if !errors.Is(err, ErrEmptyQueue) {
		t.Fatalf("Dequeue() error = %v, want %v", err, ErrEmptyQueue)
	}
}

func TestCircularQueue_PeekEmpty(t *testing.T) {
	q := &CircularQueue[int]{
		data: make([]int, 2),
		cap:  2,
	}

	_, err := q.Peek()
	if !errors.Is(err, ErrEmptyQueue) {
		t.Fatalf("Peek() error = %v, want %v", err, ErrEmptyQueue)
	}
}

func TestCircularQueue_EnqueueFull(t *testing.T) {
	q := &CircularQueue[int]{
		data: make([]int, 2),
		cap:  2,
	}

	if err := q.Enqueue(1); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if err := q.Enqueue(2); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	if !q.IsFull() {
		t.Fatal("expected queue to be full")
	}

	err := q.Enqueue(3)
	if !errors.Is(err, ErrFullQueue) {
		t.Fatalf("Enqueue() error = %v, want %v", err, ErrFullQueue)
	}
}

func TestCircularQueue_Wraparound(t *testing.T) {
	q := &CircularQueue[int]{
		data: make([]int, 3),
		cap:  3,
	}

	for _, value := range []int{1, 2, 3} {
		if err := q.Enqueue(value); err != nil {
			t.Fatalf("Enqueue(%d) error = %v", value, err)
		}
	}

	got, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}
	if got != 1 {
		t.Fatalf("Dequeue() = %d, want 1", got)
	}

	if err := q.Enqueue(4); err != nil {
		t.Fatalf("Enqueue() after wraparound error = %v", err)
	}

	want := []int{2, 3, 4}
	for _, expected := range want {
		got, err := q.Dequeue()
		if err != nil {
			t.Fatalf("Dequeue() error = %v", err)
		}
		if got != expected {
			t.Fatalf("Dequeue() = %d, want %d", got, expected)
		}
	}
}

func TestCircularQueue_EnqueueOverwriteWhenNotFull(t *testing.T) {
	q := &CircularQueue[int]{
		data: make([]int, 3),
		cap:  3,
	}

	q.EnqueueOverwrite(1)
	q.EnqueueOverwrite(2)

	if q.Size() != 2 {
		t.Fatalf("Size() = %d, want 2", q.Size())
	}

	want := []int{1, 2}
	for _, expected := range want {
		got, err := q.Dequeue()
		if err != nil {
			t.Fatalf("Dequeue() error = %v", err)
		}
		if got != expected {
			t.Fatalf("Dequeue() = %d, want %d", got, expected)
		}
	}
}

func TestCircularQueue_EnqueueOverwriteWhenFull(t *testing.T) {
	q := &CircularQueue[int]{
		data: make([]int, 3),
		cap:  3,
	}

	q.EnqueueOverwrite(1)
	q.EnqueueOverwrite(2)
	q.EnqueueOverwrite(3)

	if !q.IsFull() {
		t.Fatal("expected queue to be full")
	}

	q.EnqueueOverwrite(4)

	if q.Size() != 3 {
		t.Fatalf("Size() = %d, want 3", q.Size())
	}

	want := []int{2, 3, 4}
	for _, expected := range want {
		got, err := q.Dequeue()
		if err != nil {
			t.Fatalf("Dequeue() error = %v", err)
		}
		if got != expected {
			t.Fatalf("Dequeue() = %d, want %d", got, expected)
		}
	}
}

func TestCircularQueue_GenericStruct(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	q := &CircularQueue[item]{
		data: make([]item, 2),
		cap:  2,
	}

	expected := item{id: 1, name: "sensor"}

	if err := q.Enqueue(expected); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	got, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}

	if got != expected {
		t.Fatalf("Dequeue() = %+v, want %+v", got, expected)
	}
}

func TestCircularQueue_ImplementsQueueInterface(t *testing.T) {
	var _ Queue[int] = (*CircularQueue[int])(nil)
}
