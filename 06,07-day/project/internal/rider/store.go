package rider

import (
	"strings"
	"sync"

	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
)

type MemoryStore struct {
	mu         sync.RWMutex
	ridersByID map[string]*models.Rider
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		ridersByID: make(map[string]*models.Rider),
	}
}

func (s *MemoryStore) Register(rider *models.Rider) error {
	if !models.IsValidRider(rider) {
		return appErrors.ErrInvalidRider
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.ridersByID[rider.ID]; exists {
		return appErrors.ErrDuplicateRider
	}

	s.ridersByID[rider.ID] = rider

	return nil
}

func (s *MemoryStore) Get(riderID string) (*models.Rider, error) {
	if strings.TrimSpace(riderID) == "" {
		return nil, appErrors.ErrRiderNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rider, exists := s.ridersByID[riderID]
	if !exists {
		return nil, appErrors.ErrRiderNotFound
	}

	return rider, nil
}

func (s *MemoryStore) UpdateLocation(riderID string, loc models.Location) error {
	if strings.TrimSpace(riderID) == "" {
		return appErrors.ErrRiderNotFound
	}

	if !models.IsValidLocation(loc) {
		return appErrors.ErrInvalidLocation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	rider, exists := s.ridersByID[riderID]
	if !exists {
		return appErrors.ErrRiderNotFound
	}

	rider.Location = loc

	return nil
}

func (s *MemoryStore) AddRideToHistory(riderID string, rideID string) error {
	if strings.TrimSpace(riderID) == "" {
		return appErrors.ErrRiderNotFound
	}

	if strings.TrimSpace(rideID) == "" {
		return appErrors.ErrInvalidRide
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	rider, exists := s.ridersByID[riderID]
	if !exists {
		return appErrors.ErrRiderNotFound
	}

	rider.RideHistory = append(rider.RideHistory, rideID)

	return nil
}
