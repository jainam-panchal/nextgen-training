package queue

type Queue[T any] interface {
	Enqueue(T) error
	EnqueueOverwrite(T)
	Dequeue() (T, error)
	Peek() (T, error)
	IsFull() bool
	IsEmpty() bool
	Size() int
	Len() int
	Cap() int
}
