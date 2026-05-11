package rides

import (
	"strings"
	"sync"

	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
)

type MemoryRideHistoryStore struct {
	mu        sync.RWMutex
	ridesByID map[string]*models.Ride
}

func NewMemoryRideHistoryStore() *MemoryRideHistoryStore {
	return &MemoryRideHistoryStore{
		ridesByID: make(map[string]*models.Ride),
	}
}

func (s *MemoryRideHistoryStore) Save(ride *models.Ride) error {
	if !models.IsValidRide(ride) {
		return appErrors.ErrInvalidRide
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.ridesByID[ride.ID]; exists {
		return appErrors.ErrDuplicateRide
	}

	s.ridesByID[ride.ID] = ride

	return nil
}

func (s *MemoryRideHistoryStore) Get(rideID string) (*models.Ride, error) {
	if strings.TrimSpace(rideID) == "" {
		return nil, appErrors.ErrRideNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ride, exists := s.ridesByID[rideID]
	if !exists {
		return nil, appErrors.ErrRideNotFound
	}

	return ride, nil
}

func (s *MemoryRideHistoryStore) List() []*models.Ride {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rides := make([]*models.Ride, 0, len(s.ridesByID))

	for _, ride := range s.ridesByID {
		rides = append(rides, ride)
	}

	return rides
}

func (s *MemoryRideHistoryStore) ListByDriver(driverID string) []*models.Ride {
	if strings.TrimSpace(driverID) == "" {
		return []*models.Ride{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rides := make([]*models.Ride, 0)

	for _, ride := range s.ridesByID {
		if ride.DriverID == driverID {
			rides = append(rides, ride)
		}
	}

	return rides
}
