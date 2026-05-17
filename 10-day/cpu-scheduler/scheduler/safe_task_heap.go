package scheduler

import (
	"sync"

	"cpu-scheduler/heap"
)

type SafeTaskHeap struct {
	mutex sync.Mutex
	heap  *heap.MinHeap[*Task]
	notifyCh chan struct{}
}

func NewSafeTaskHeap() *SafeTaskHeap {
	return &SafeTaskHeap{
		heap:     heap.NewMinHeap[*Task](LessTask),
		notifyCh: make(chan struct{}, 1),
	}
}

func (q *SafeTaskHeap) Push(task *Task) {
	q.mutex.Lock()
	q.heap.Push(task)
	q.mutex.Unlock()

	select {
	case q.notifyCh <- struct{}{}:
	default:
	}
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

func (q *SafeTaskHeap) NotifyChan() <-chan struct{} {
	return q.notifyCh
}

func (q *SafeTaskHeap) AgeWaitingTasks() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	tasks := make([]*Task, 0, q.heap.Len())

	for !q.heap.IsEmpty() {
		task, err := q.heap.Pop()
		if err != nil {
			break
		}

		if task.Priority > 1 {
			task.Priority--
		}

		tasks = append(tasks, task)
	}

	for _, task := range tasks {
		q.heap.Push(task)
	}
}
