package driver

import (
	stderrors "errors"
	"testing"

	"ride-sharing/internal/config"
	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/geo"
	"ride-sharing/internal/models"
)

func newValidDriver(name string, lat float64, lng float64) *models.Driver {
	return &models.Driver{
		ID:       models.NewDriverID(),
		Name:     name,
		Location: models.Location{Lat: lat, Lng: lng},
		Status:   models.DriverOffline,
		Rating:   4.8,
	}
}

// Tests registration validation and duplicate handling.
func TestRegister(t *testing.T) {
	store := NewMemoryStore(config.DefaultConfig())

	driver := newValidDriver("Jainam", 19.08, 72.88)
	if err := store.Register(driver); err != nil {
		t.Fatalf("register should succeed: %v", err)
	}

	storedDriver, err := store.Get(driver.ID)
	if err != nil {
		t.Fatalf("get after register failed: %v", err)
	}
	if storedDriver.Status != models.DriverOffline {
		t.Fatalf("expected status offline, got %s", storedDriver.Status)
	}

	if err := store.Register(driver); !stderrors.Is(err, appErrors.ErrDuplicateDriver) {
		t.Fatalf("expected ErrDuplicateDriver, got %v", err)
	}

	invalidDriver := newValidDriver("", 19.08, 72.88)
	if err := store.Register(invalidDriver); !stderrors.Is(err, appErrors.ErrInvalidDriver) {
		t.Fatalf("expected ErrInvalidDriver, got %v", err)
	}
}

// Tests online/offline transitions and idempotency.
func TestGoOnlineGoOfflineTransitions(t *testing.T) {
	store := NewMemoryStore(config.DefaultConfig())
	driver := newValidDriver("Jainam", 19.08, 72.88)
	_ = store.Register(driver)

	if err := store.GoOnline(driver.ID); err != nil {
		t.Fatalf("go online should succeed: %v", err)
	}
	if err := store.GoOnline(driver.ID); err != nil {
		t.Fatalf("go online should be idempotent: %v", err)
	}

	updatedDriver, _ := store.Get(driver.ID)
	if updatedDriver.Status != models.DriverAvailable {
		t.Fatalf("expected status available, got %s", updatedDriver.Status)
	}

	if err := store.GoOffline(driver.ID); err != nil {
		t.Fatalf("go offline should succeed: %v", err)
	}
	if err := store.GoOffline(driver.ID); err != nil {
		t.Fatalf("go offline should be idempotent: %v", err)
	}

	updatedDriver, _ = store.Get(driver.ID)
	if updatedDriver.Status != models.DriverOffline {
		t.Fatalf("expected status offline, got %s", updatedDriver.Status)
	}

	if err := store.GoOnline("UNKNOWN"); !stderrors.Is(err, appErrors.ErrDriverNotFound) {
		t.Fatalf("expected ErrDriverNotFound, got %v", err)
	}
	if err := store.GoOffline("UNKNOWN"); !stderrors.Is(err, appErrors.ErrDriverNotFound) {
		t.Fatalf("expected ErrDriverNotFound, got %v", err)
	}
}

// Tests assign/release transitions and invalid state transitions.
func TestAssignReleaseTransitions(t *testing.T) {
	store := NewMemoryStore(config.DefaultConfig())
	driver := newValidDriver("Jainam", 19.08, 72.88)
	_ = store.Register(driver)
	_ = store.GoOnline(driver.ID)

	if err := store.Assign(driver.ID); err != nil {
		t.Fatalf("assign should succeed: %v", err)
	}
	if err := store.Assign(driver.ID); err != nil {
		t.Fatalf("assign should be idempotent for busy driver: %v", err)
	}

	updatedDriver, _ := store.Get(driver.ID)
	if updatedDriver.Status != models.DriverBusy {
		t.Fatalf("expected status busy, got %s", updatedDriver.Status)
	}

	if err := store.Release(driver.ID); err != nil {
		t.Fatalf("release should succeed: %v", err)
	}
	if err := store.Release(driver.ID); err != nil {
		t.Fatalf("release should be idempotent for available driver: %v", err)
	}

	updatedDriver, _ = store.Get(driver.ID)
	if updatedDriver.Status != models.DriverAvailable {
		t.Fatalf("expected status available, got %s", updatedDriver.Status)
	}

	offlineDriver := newValidDriver("Ravi", 19.08, 72.88)
	_ = store.Register(offlineDriver)
	if err := store.Assign(offlineDriver.ID); !stderrors.Is(err, appErrors.ErrInvalidStatusChange) {
		t.Fatalf("expected ErrInvalidStatusChange on assign offline, got %v", err)
	}
	if err := store.Release(offlineDriver.ID); !stderrors.Is(err, appErrors.ErrInvalidStatusChange) {
		t.Fatalf("expected ErrInvalidStatusChange on release offline, got %v", err)
	}
}

// Tests location updates and location-dependent indexing behavior.
func TestUpdateLocation(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMemoryStore(cfg)
	driver := newValidDriver("Jainam", 19.08, 72.88)
	_ = store.Register(driver)
	_ = store.GoOnline(driver.ID)

	if err := store.UpdateLocation(driver.ID, models.Location{Lat: 200, Lng: 72.88}); !stderrors.Is(err, appErrors.ErrInvalidLocation) {
		t.Fatalf("expected ErrInvalidLocation, got %v", err)
	}
	if err := store.UpdateLocation("UNKNOWN", models.Location{Lat: 19.08, Lng: 72.88}); !stderrors.Is(err, appErrors.ErrDriverNotFound) {
		t.Fatalf("expected ErrDriverNotFound, got %v", err)
	}

	newLoc := models.Location{Lat: 19.09, Lng: 72.89}
	if err := store.UpdateLocation(driver.ID, newLoc); err != nil {
		t.Fatalf("update location should succeed: %v", err)
	}

	updatedDriver, err := store.Get(driver.ID)
	if err != nil {
		t.Fatalf("get driver failed: %v", err)
	}
	if updatedDriver.Location != newLoc {
		t.Fatalf("expected location %+v, got %+v", newLoc, updatedDriver.Location)
	}

	expectedBlock := geo.GetBlockID(newLoc, cfg.CellSizeDegrees)
	if updatedDriver.BlockID != expectedBlock {
		t.Fatalf("expected block %+v, got %+v", expectedBlock, updatedDriver.BlockID)
	}

	foundDriver, _, err := store.FindNearestAvailable(newLoc, 5)
	if err != nil {
		t.Fatalf("find nearest failed: %v", err)
	}
	if foundDriver.ID != driver.ID {
		t.Fatalf("expected driver %s, got %s", driver.ID, foundDriver.ID)
	}
}

// Tests nearest available driver selection and input validation.
func TestFindNearestAvailable(t *testing.T) {
	store := NewMemoryStore(config.DefaultConfig())
	query := models.Location{Lat: 19.08, Lng: 72.88}

	if _, _, err := store.FindNearestAvailable(query, 5); !stderrors.Is(err, appErrors.ErrNoDriverFound) {
		t.Fatalf("expected ErrNoDriverFound, got %v", err)
	}

	nearDriver := newValidDriver("Near", 19.081, 72.881)
	farDriver := newValidDriver("Far", 19.20, 73.00)
	_ = store.Register(nearDriver)
	_ = store.Register(farDriver)
	_ = store.GoOnline(nearDriver.ID)
	_ = store.GoOnline(farDriver.ID)

	foundDriver, distanceKm, err := store.FindNearestAvailable(query, 5)
	if err != nil {
		t.Fatalf("find nearest should succeed: %v", err)
	}
	if foundDriver.ID != nearDriver.ID {
		t.Fatalf("expected nearest driver %s, got %s", nearDriver.ID, foundDriver.ID)
	}
	if distanceKm <= 0 || distanceKm > 5 {
		t.Fatalf("expected distance in (0,5], got %f", distanceKm)
	}

	if _, _, err := store.FindNearestAvailable(models.Location{Lat: 200, Lng: 72.88}, 5); !stderrors.Is(err, appErrors.ErrInvalidLocation) {
		t.Fatalf("expected ErrInvalidLocation, got %v", err)
	}
	if _, _, err := store.FindNearestAvailable(query, 0); !stderrors.Is(err, appErrors.ErrNoDriverFound) {
		t.Fatalf("expected ErrNoDriverFound for non-positive max distance, got %v", err)
	}
}

// Tests equal-distance tie-break rule (lower ID wins).
func TestFindNearestAvailableTieBreakByID(t *testing.T) {
	store := NewMemoryStore(config.DefaultConfig())
	query := models.Location{Lat: 19.08, Lng: 72.88}

	driverA := &models.Driver{
		ID:       "DRV-000001",
		Name:     "A",
		Location: models.Location{Lat: 19.08, Lng: 72.88},
		Status:   models.DriverOffline,
		Rating:   4.5,
	}
	driverB := &models.Driver{
		ID:       "DRV-000002",
		Name:     "B",
		Location: models.Location{Lat: 19.08, Lng: 72.88},
		Status:   models.DriverOffline,
		Rating:   4.5,
	}

	_ = store.Register(driverA)
	_ = store.Register(driverB)
	_ = store.GoOnline(driverA.ID)
	_ = store.GoOnline(driverB.ID)

	foundDriver, _, err := store.FindNearestAvailable(query, 5)
	if err != nil {
		t.Fatalf("find nearest should succeed: %v", err)
	}

	if foundDriver.ID != driverA.ID {
		t.Fatalf("expected tie-break winner %s, got %s", driverA.ID, foundDriver.ID)
	}
}

// Tests top-N nearest selection, limits, and deterministic ordering.
func TestFindNNearestAvailable(t *testing.T) {
	store := NewMemoryStore(config.DefaultConfig())
	query := models.Location{Lat: 19.08, Lng: 72.88}

	drivers, err := store.FindNNearestAvailable(query, 0, 5)
	if err != nil {
		t.Fatalf("limit<=0 should not error: %v", err)
	}
	if len(drivers) != 0 {
		t.Fatalf("expected empty list for limit<=0, got %d", len(drivers))
	}

	if _, err := store.FindNNearestAvailable(query, 2, 5); !stderrors.Is(err, appErrors.ErrNoDriverFound) {
		t.Fatalf("expected ErrNoDriverFound, got %v", err)
	}

	driverA := newValidDriver("A", 19.0805, 72.88)
	driverB := newValidDriver("B", 19.0810, 72.88)
	driverC := newValidDriver("C", 19.0810, 72.88)
	driverFar := newValidDriver("Far", 19.20, 73.00)
	driverB.ID = "DRV-000002"
	driverC.ID = "DRV-000001"

	_ = store.Register(driverA)
	_ = store.Register(driverB)
	_ = store.Register(driverC)
	_ = store.Register(driverFar)
	_ = store.GoOnline(driverA.ID)
	_ = store.GoOnline(driverB.ID)
	_ = store.GoOnline(driverC.ID)
	_ = store.GoOnline(driverFar.ID)

	nearestTwo, err := store.FindNNearestAvailable(query, 2, 5)
	if err != nil {
		t.Fatalf("find N nearest should succeed: %v", err)
	}
	if len(nearestTwo) != 2 {
		t.Fatalf("expected 2 drivers, got %d", len(nearestTwo))
	}
	if nearestTwo[0].ID != driverA.ID {
		t.Fatalf("expected nearest first %s, got %s", driverA.ID, nearestTwo[0].ID)
	}

	allInRange, err := store.FindNNearestAvailable(query, 10, 5)
	if err != nil {
		t.Fatalf("find all in range should succeed: %v", err)
	}
	if len(allInRange) != 3 {
		t.Fatalf("expected 3 in-range drivers, got %d", len(allInRange))
	}

	// B and C are at equal distance, so lower ID must come first.
	if allInRange[1].ID != "DRV-000001" || allInRange[2].ID != "DRV-000002" {
		t.Fatalf("expected tie-break order DRV-000001 then DRV-000002, got %s then %s", allInRange[1].ID, allInRange[2].ID)
	}
}

// Tests ride history append and validation for invalid IDs.
func TestAddRideToHistory(t *testing.T) {
	store := NewMemoryStore(config.DefaultConfig())
	driver := newValidDriver("Jainam", 19.08, 72.88)
	_ = store.Register(driver)

	if err := store.AddRideToHistory(driver.ID, "RID-000001"); err != nil {
		t.Fatalf("add ride history should succeed: %v", err)
	}

	updatedDriver, _ := store.Get(driver.ID)
	if len(updatedDriver.RideHistory) != 1 || updatedDriver.RideHistory[0] != "RID-000001" {
		t.Fatalf("expected ride history to contain RID-000001")
	}

	if err := store.AddRideToHistory("", "RID-000002"); !stderrors.Is(err, appErrors.ErrDriverNotFound) {
		t.Fatalf("expected ErrDriverNotFound for empty driver ID, got %v", err)
	}
	if err := store.AddRideToHistory(driver.ID, ""); !stderrors.Is(err, appErrors.ErrInvalidRide) {
		t.Fatalf("expected ErrInvalidRide for empty ride ID, got %v", err)
	}
	if err := store.AddRideToHistory("UNKNOWN", "RID-000002"); !stderrors.Is(err, appErrors.ErrDriverNotFound) {
		t.Fatalf("expected ErrDriverNotFound for unknown driver ID, got %v", err)
	}
}
