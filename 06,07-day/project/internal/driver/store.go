package driver

import (
	"sort"
	"strings"
	"sync"

	"ride-sharing/internal/config"
	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/geo"
	"ride-sharing/internal/models"
)

type MemoryStore struct {
	mu                      sync.RWMutex
	cfg                     config.Config
	driversByID             map[string]*models.Driver
	availableDriversByBlock map[models.BlockID]map[string]*models.Driver
}

func NewMemoryStore(cfg config.Config) *MemoryStore {
	return &MemoryStore{
		cfg:                     cfg,
		driversByID:             make(map[string]*models.Driver),
		availableDriversByBlock: make(map[models.BlockID]map[string]*models.Driver),
	}
}

func (s *MemoryStore) Register(driver *models.Driver) error {
	if !models.IsValidDriver(driver) {
		return appErrors.ErrInvalidDriver
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.driversByID[driver.ID]; exists {
		return appErrors.ErrDuplicateDriver
	}

	driver.Status = models.DriverOffline
	driver.BlockID = geo.GetBlockID(driver.Location, s.cfg.CellSizeDegrees)

	s.driversByID[driver.ID] = driver

	return nil
}

func (s *MemoryStore) Get(driverID string) (*models.Driver, error) {
	if strings.TrimSpace(driverID) == "" {
		return nil, appErrors.ErrDriverNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	driver, exists := s.driversByID[driverID]
	if !exists {
		return nil, appErrors.ErrDriverNotFound
	}

	return driver, nil
}

func (s *MemoryStore) GoOnline(driverID string) error {
	if strings.TrimSpace(driverID) == "" {
		return appErrors.ErrDriverNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.driversByID[driverID]
	if !exists {
		return appErrors.ErrDriverNotFound
	}

	switch driver.Status {
	case models.DriverAvailable:
		return nil

	case models.DriverOffline:
		driver.Status = models.DriverAvailable
		driver.BlockID = geo.GetBlockID(driver.Location, s.cfg.CellSizeDegrees)
		s.addAvailableDriverLocked(driver)
		return nil

	case models.DriverBusy:
		return appErrors.ErrInvalidStatusChange

	default:
		return appErrors.ErrInvalidStatus
	}
}

func (s *MemoryStore) GoOffline(driverID string) error {
	if strings.TrimSpace(driverID) == "" {
		return appErrors.ErrDriverNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.driversByID[driverID]
	if !exists {
		return appErrors.ErrDriverNotFound
	}

	switch driver.Status {
	case models.DriverOffline:
		return nil

	case models.DriverAvailable:
		s.removeAvailableDriverLocked(driver)
		driver.Status = models.DriverOffline
		return nil

	case models.DriverBusy:
		return appErrors.ErrInvalidStatusChange

	default:
		return appErrors.ErrInvalidStatus
	}
}

func (s *MemoryStore) Assign(driverID string) error {
	if strings.TrimSpace(driverID) == "" {
		return appErrors.ErrDriverNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.driversByID[driverID]
	if !exists {
		return appErrors.ErrDriverNotFound
	}

	switch driver.Status {
	case models.DriverBusy:
		return nil

	case models.DriverAvailable:
		s.removeAvailableDriverLocked(driver)
		driver.Status = models.DriverBusy
		return nil

	case models.DriverOffline:
		return appErrors.ErrInvalidStatusChange

	default:
		return appErrors.ErrInvalidStatus
	}
}

func (s *MemoryStore) Release(driverID string) error {
	if strings.TrimSpace(driverID) == "" {
		return appErrors.ErrDriverNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.driversByID[driverID]
	if !exists {
		return appErrors.ErrDriverNotFound
	}

	switch driver.Status {
	case models.DriverAvailable:
		return nil

	case models.DriverBusy:
		driver.Status = models.DriverAvailable
		driver.BlockID = geo.GetBlockID(driver.Location, s.cfg.CellSizeDegrees)
		s.addAvailableDriverLocked(driver)
		return nil

	case models.DriverOffline:
		return appErrors.ErrInvalidStatusChange

	default:
		return appErrors.ErrInvalidStatus
	}
}

func (s *MemoryStore) UpdateLocation(driverID string, loc models.Location) error {
	if strings.TrimSpace(driverID) == "" {
		return appErrors.ErrDriverNotFound
	}

	if !models.IsValidLocation(loc) {
		return appErrors.ErrInvalidLocation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.driversByID[driverID]
	if !exists {
		return appErrors.ErrDriverNotFound
	}

	if driver.Status == models.DriverAvailable {
		s.removeAvailableDriverLocked(driver)
	}

	driver.Location = loc
	driver.BlockID = geo.GetBlockID(loc, s.cfg.CellSizeDegrees)

	if driver.Status == models.DriverAvailable {
		s.addAvailableDriverLocked(driver)
	}

	return nil
}

func (s *MemoryStore) FindNearestAvailable(
	loc models.Location,
	maxDistanceKm float64,
) (*models.Driver, float64, error) {
	if !models.IsValidLocation(loc) {
		return nil, 0, appErrors.ErrInvalidLocation
	}

	if maxDistanceKm <= 0 {
		return nil, 0, appErrors.ErrNoDriverFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	centerBlock := geo.GetBlockID(loc, s.cfg.CellSizeDegrees)
	searchRing := s.cfg.NeighborRing()

	var nearestDriver *models.Driver
	nearestDistanceKm := maxDistanceKm

	for _, block := range geo.NeighborBlocks(centerBlock, searchRing) {
		for _, driver := range s.availableDriversByBlock[block] {
			distanceKm := geo.DistanceKm(loc, driver.Location)

			if distanceKm > maxDistanceKm {
				continue
			}

			if nearestDriver == nil ||
				distanceKm < nearestDistanceKm ||
				(distanceKm == nearestDistanceKm && driver.ID < nearestDriver.ID) {
				nearestDriver = driver
				nearestDistanceKm = distanceKm
			}
		}
	}

	if nearestDriver == nil {
		return nil, 0, appErrors.ErrNoDriverFound
	}

	return nearestDriver, nearestDistanceKm, nil
}

func (s *MemoryStore) FindNNearestAvailable(
	loc models.Location,
	limit int,
	maxDistanceKm float64,
) ([]*models.Driver, error) {
	if !models.IsValidLocation(loc) {
		return nil, appErrors.ErrInvalidLocation
	}

	if limit <= 0 {
		return []*models.Driver{}, nil
	}

	if maxDistanceKm <= 0 {
		return nil, appErrors.ErrNoDriverFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	type driverDistanceCandidate struct {
		driver     *models.Driver
		distanceKm float64
	}

	centerBlock := geo.GetBlockID(loc, s.cfg.CellSizeDegrees)
	searchRing := s.cfg.NeighborRing()

	candidates := make([]driverDistanceCandidate, 0)

	for _, block := range geo.NeighborBlocks(centerBlock, searchRing) {
		for _, driver := range s.availableDriversByBlock[block] {
			distanceKm := geo.DistanceKm(loc, driver.Location)

			if distanceKm <= maxDistanceKm {
				candidates = append(candidates, driverDistanceCandidate{
					driver:     driver,
					distanceKm: distanceKm,
				})
			}
		}
	}

	if len(candidates) == 0 {
		return nil, appErrors.ErrNoDriverFound
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distanceKm == candidates[j].distanceKm {
			return candidates[i].driver.ID < candidates[j].driver.ID
		}

		return candidates[i].distanceKm < candidates[j].distanceKm
	})

	if limit > len(candidates) {
		limit = len(candidates)
	}

	nearestDrivers := make([]*models.Driver, 0, limit)
	for i := 0; i < limit; i++ {
		nearestDrivers = append(nearestDrivers, candidates[i].driver)
	}

	return nearestDrivers, nil
}

func (s *MemoryStore) addAvailableDriverLocked(driver *models.Driver) {
	if s.availableDriversByBlock[driver.BlockID] == nil {
		s.availableDriversByBlock[driver.BlockID] = make(map[string]*models.Driver)
	}

	s.availableDriversByBlock[driver.BlockID][driver.ID] = driver
}

func (s *MemoryStore) removeAvailableDriverLocked(driver *models.Driver) {
	driversInBlock, exists := s.availableDriversByBlock[driver.BlockID]
	if !exists {
		return
	}

	delete(driversInBlock, driver.ID)

	if len(driversInBlock) == 0 {
		delete(s.availableDriversByBlock, driver.BlockID)
	}
}

func (s *MemoryStore) AddRideToHistory(driverID string, rideID string) error {
	if strings.TrimSpace(driverID) == "" {
		return appErrors.ErrDriverNotFound
	}

	if strings.TrimSpace(rideID) == "" {
		return appErrors.ErrInvalidRide
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.driversByID[driverID]
	if !exists {
		return appErrors.ErrDriverNotFound
	}

	driver.RideHistory = append(driver.RideHistory, rideID)

	return nil
}
