package ridequeue

import (
	"sync"

	"ride-sharing/internal/ds/heap"
	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
)

type RideRequestPriorityQueue struct {
	mu       sync.RWMutex
	requests *heap.MinHeap[*models.RideRequest]
}

func NewRideRequestPriorityQueue() *RideRequestPriorityQueue {
	return &RideRequestPriorityQueue{
		requests: heap.NewMinHeap(func(
			firstRequest *models.RideRequest,
			secondRequest *models.RideRequest,
		) bool {
			return firstRequest.RequestTime.Before(secondRequest.RequestTime)
		}),
	}
}

func (q *RideRequestPriorityQueue) Push(request *models.RideRequest) error {
	if !models.IsValidRideRequest(request) {
		return appErrors.ErrInvalidRideRequest
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.requests.Push(request)

	return nil
}

func (q *RideRequestPriorityQueue) Peek() (*models.RideRequest, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.requests.Peek()
}

func (q *RideRequestPriorityQueue) Pop() (*models.RideRequest, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.requests.Pop()
}

func (q *RideRequestPriorityQueue) Snapshot() []*models.RideRequest {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.requests.Items()
}

func (q *RideRequestPriorityQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.requests.Len()
}

func (q *RideRequestPriorityQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.requests.IsEmpty()
}
