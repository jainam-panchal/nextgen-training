package ridequeue

import (
	stderrors "errors"
	"testing"
	"time"

	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
)

func newRideRequest(riderID string, at time.Time) *models.RideRequest {
	return &models.RideRequest{
		ID:          models.NewRequestID(),
		RiderID:     riderID,
		Pickup:      models.Location{Lat: 19.08, Lng: 72.88},
		Dropoff:     models.Location{Lat: 19.09, Lng: 72.89},
		RequestTime: at,
	}
}

// Tests queue ordering by oldest request time first.
func TestRideRequestPriorityQueueOrder(t *testing.T) {
	q := NewRideRequestPriorityQueue()
	base := time.Now()

	r3 := newRideRequest("R3", base.Add(3*time.Minute))
	r1 := newRideRequest("R1", base.Add(1*time.Minute))
	r2 := newRideRequest("R2", base.Add(2*time.Minute))

	if err := q.Push(r3); err != nil {
		t.Fatalf("push r3 failed: %v", err)
	}
	if err := q.Push(r1); err != nil {
		t.Fatalf("push r1 failed: %v", err)
	}
	if err := q.Push(r2); err != nil {
		t.Fatalf("push r2 failed: %v", err)
	}

	peeked, err := q.Peek()
	if err != nil {
		t.Fatalf("peek failed: %v", err)
	}
	if peeked.ID != r1.ID {
		t.Fatalf("expected oldest request first, got %s", peeked.ID)
	}

	first, _ := q.Pop()
	second, _ := q.Pop()
	third, _ := q.Pop()

	if first.ID != r1.ID || second.ID != r2.ID || third.ID != r3.ID {
		t.Fatalf("unexpected pop order: %s, %s, %s", first.ID, second.ID, third.ID)
	}
}

// Tests invalid request validation and empty queue errors.
func TestRideRequestPriorityQueueErrors(t *testing.T) {
	q := NewRideRequestPriorityQueue()

	if err := q.Push(&models.RideRequest{}); !stderrors.Is(err, appErrors.ErrInvalidRideRequest) {
		t.Fatalf("expected ErrInvalidRideRequest, got %v", err)
	}

	if _, err := q.Peek(); !stderrors.Is(err, appErrors.ErrEmptyQueue) {
		t.Fatalf("expected ErrEmptyQueue from peek, got %v", err)
	}
	if _, err := q.Pop(); !stderrors.Is(err, appErrors.ErrEmptyQueue) {
		t.Fatalf("expected ErrEmptyQueue from pop, got %v", err)
	}
}

// Tests snapshot behavior: returns copy and does not mutate queue.
func TestRideRequestPriorityQueueSnapshot(t *testing.T) {
	q := NewRideRequestPriorityQueue()
	base := time.Now()
	r1 := newRideRequest("R1", base)
	r2 := newRideRequest("R2", base.Add(time.Minute))

	_ = q.Push(r1)
	_ = q.Push(r2)

	snapshot := q.Snapshot()
	if len(snapshot) != 2 {
		t.Fatalf("expected snapshot size 2, got %d", len(snapshot))
	}
	if q.Len() != 2 {
		t.Fatalf("snapshot should not change queue length")
	}

	snapshot[0] = nil

	peeked, err := q.Peek()
	if err != nil {
		t.Fatalf("peek failed after snapshot mutation: %v", err)
	}
	if peeked == nil {
		t.Fatalf("queue should remain intact after snapshot mutation")
	}
}
