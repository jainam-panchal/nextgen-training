package rides

import (
	"strings"
	"sync"

	"ride-sharing/internal/ds/linkedlist"
	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
)

type MemoryActiveRideTracker struct {
	mu                  sync.RWMutex
	activeRides         *linkedlist.DoublyLinkedList[*models.Ride]
	activeRideNodesByID map[string]*linkedlist.Node[*models.Ride]
}

func NewMemoryActiveRideTracker() *MemoryActiveRideTracker {
	return &MemoryActiveRideTracker{
		activeRides:         linkedlist.NewDoublyLinkedList[*models.Ride](),
		activeRideNodesByID: make(map[string]*linkedlist.Node[*models.Ride]),
	}
}

func (t *MemoryActiveRideTracker) Add(ride *models.Ride) error {
	if !models.IsValidRide(ride) {
		return appErrors.ErrInvalidRide
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.activeRideNodesByID[ride.ID]; exists {
		return appErrors.ErrDuplicateRide
	}

	node := t.activeRides.PushBack(ride)
	t.activeRideNodesByID[ride.ID] = node

	return nil
}

func (t *MemoryActiveRideTracker) Get(rideID string) (*models.Ride, error) {
	if strings.TrimSpace(rideID) == "" {
		return nil, appErrors.ErrRideNotFound
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	node, exists := t.activeRideNodesByID[rideID]
	if !exists {
		return nil, appErrors.ErrRideNotFound
	}

	return node.Value(), nil
}

func (t *MemoryActiveRideTracker) Remove(rideID string) (*models.Ride, error) {
	if strings.TrimSpace(rideID) == "" {
		return nil, appErrors.ErrRideNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	node, exists := t.activeRideNodesByID[rideID]
	if !exists {
		return nil, appErrors.ErrRideNotFound
	}

	removedRide, err := t.activeRides.Remove(node)
	if err != nil {
		return nil, err
	}

	delete(t.activeRideNodesByID, rideID)

	return removedRide, nil
}

func (t *MemoryActiveRideTracker) List() []*models.Ride {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.activeRides.Values()
}

func (t *MemoryActiveRideTracker) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.activeRideNodesByID)
}

func (t *MemoryActiveRideTracker) IsEmpty() bool {
	return t.Len() == 0
}
