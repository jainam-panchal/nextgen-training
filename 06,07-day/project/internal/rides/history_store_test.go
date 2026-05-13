package rides

import (
	stderrors "errors"
	"testing"
	"time"

	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
)

func newCompletedRide(driverID string) *models.Ride {
	now := time.Now()
	return &models.Ride{
		ID:          models.NewRideID(),
		RiderID:     models.NewRiderID(),
		DriverID:    driverID,
		Pickup:      models.Location{Lat: 19.08, Lng: 72.88},
		Dropoff:     models.Location{Lat: 19.09, Lng: 72.89},
		Status:      models.RideCompleted,
		RequestTime: now.Add(-10 * time.Minute),
		StartTime:   now.Add(-8 * time.Minute),
		EndTime:     now.Add(-1 * time.Minute),
		Fare:        110,
		DistanceKm:  4.8,
	}
}

// Tests save/get/list/list-by-driver happy flow.
func TestMemoryRideHistoryStoreBasicFlow(t *testing.T) {
	store := NewMemoryRideHistoryStore()
	driverA := models.NewDriverID()
	driverB := models.NewDriverID()

	ride1 := newCompletedRide(driverA)
	ride2 := newCompletedRide(driverA)
	ride3 := newCompletedRide(driverB)

	if err := store.Save(ride1); err != nil {
		t.Fatalf("save ride1 failed: %v", err)
	}
	if err := store.Save(ride2); err != nil {
		t.Fatalf("save ride2 failed: %v", err)
	}
	if err := store.Save(ride3); err != nil {
		t.Fatalf("save ride3 failed: %v", err)
	}

	got, err := store.Get(ride1.ID)
	if err != nil {
		t.Fatalf("get ride failed: %v", err)
	}
	if got.ID != ride1.ID {
		t.Fatalf("expected ride %s, got %s", ride1.ID, got.ID)
	}

	all := store.List()
	if len(all) != 3 {
		t.Fatalf("expected 3 rides in list, got %d", len(all))
	}

	byDriverA := store.ListByDriver(driverA)
	if len(byDriverA) != 2 {
		t.Fatalf("expected 2 rides for driver A, got %d", len(byDriverA))
	}
}

// Tests invalid/duplicate/not-found branches.
func TestMemoryRideHistoryStoreErrors(t *testing.T) {
	store := NewMemoryRideHistoryStore()
	ride := newCompletedRide(models.NewDriverID())

	if err := store.Save(&models.Ride{}); !stderrors.Is(err, appErrors.ErrInvalidRide) {
		t.Fatalf("expected ErrInvalidRide, got %v", err)
	}

	if err := store.Save(ride); err != nil {
		t.Fatalf("save valid ride failed: %v", err)
	}
	if err := store.Save(ride); !stderrors.Is(err, appErrors.ErrDuplicateRide) {
		t.Fatalf("expected ErrDuplicateRide, got %v", err)
	}

	if _, err := store.Get(""); !stderrors.Is(err, appErrors.ErrRideNotFound) {
		t.Fatalf("expected ErrRideNotFound for empty ID, got %v", err)
	}
	if _, err := store.Get("UNKNOWN"); !stderrors.Is(err, appErrors.ErrRideNotFound) {
		t.Fatalf("expected ErrRideNotFound for unknown ID, got %v", err)
	}

	emptyDriverResult := store.ListByDriver("")
	if len(emptyDriverResult) != 0 {
		t.Fatalf("expected empty list for empty driver ID")
	}
}
