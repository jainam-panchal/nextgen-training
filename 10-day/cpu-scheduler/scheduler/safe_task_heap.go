package scheduler

import (
	"cpu-scheduler/heap"
	"sync"
)

type SafeTaskHeap struct {
	mutex sync.Mutex
	heap  *heap.MinHeap[*Task]
}

func NewSafeTaskHeap() *SafeTaskHeap {
	return &SafeTaskHeap{
		heap: heap.NewMinHeap[*Task](LessTask),
	}
}

func (q *SafeTaskHeap) Push(task *Task) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.heap.Push(task)
}

func (q *SafeTaskHeap) Pop() (*Task, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.heap.Pop()
}

func (q *SafeTaskHeap) Peek() (*Task, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.heap.Peek()
}

func (q *SafeTaskHeap) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.heap.Len()
}

func (q *SafeTaskHeap) IsEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.heap.IsEmpty()
}
