package queue

import "jainamp-panchal/nextgen-training/packet-router/internal/packet"

type Queue[T any] interface {
	Enqueue(value T)
	Dequeue() (T, bool)
	Peek() (T, bool)
	Len() int
}

type PacketQueue interface {
	Queue[*packet.Packet]
}

type NewQueueFunc func() PacketQueue
