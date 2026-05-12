package rides

import (
	stderrors "errors"
	"testing"
	"time"

	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
)

func newValidRide() *models.Ride {
	now := time.Now()
	return &models.Ride{
		ID:          models.NewRideID(),
		RiderID:     models.NewRiderID(),
		DriverID:    models.NewDriverID(),
		Pickup:      models.Location{Lat: 19.08, Lng: 72.88},
		Dropoff:     models.Location{Lat: 19.09, Lng: 72.89},
		Status:      models.RideAssigned,
		RequestTime: now.Add(-2 * time.Minute),
		StartTime:   now.Add(-time.Minute),
		Fare:        120,
		DistanceKm:  5,
	}
}

// Tests add/get/remove/list/len behavior for active rides.
func TestMemoryActiveRideTrackerBasicFlow(t *testing.T) {
	tracker := NewMemoryActiveRideTracker()
	ride := newValidRide()

	if err := tracker.Add(ride); err != nil {
		t.Fatalf("add ride failed: %v", err)
	}
	if tracker.Len() != 1 || tracker.IsEmpty() {
		t.Fatalf("expected one active ride")
	}

	gotRide, err := tracker.Get(ride.ID)
	if err != nil {
		t.Fatalf("get ride failed: %v", err)
	}
	if gotRide.ID != ride.ID {
		t.Fatalf("expected ride %s, got %s", ride.ID, gotRide.ID)
	}

	list := tracker.List()
	if len(list) != 1 || list[0].ID != ride.ID {
		t.Fatalf("expected list with one ride %s", ride.ID)
	}

	removed, err := tracker.Remove(ride.ID)
	if err != nil {
		t.Fatalf("remove ride failed: %v", err)
	}
	if removed.ID != ride.ID {
		t.Fatalf("expected removed ride %s, got %s", ride.ID, removed.ID)
	}
	if tracker.Len() != 0 || !tracker.IsEmpty() {
		t.Fatalf("expected tracker to be empty")
	}
}

// Tests duplicate/add validation and not-found removal paths.
func TestMemoryActiveRideTrackerErrors(t *testing.T) {
	tracker := NewMemoryActiveRideTracker()
	ride := newValidRide()

	if err := tracker.Add(&models.Ride{}); !stderrors.Is(err, appErrors.ErrInvalidRide) {
		t.Fatalf("expected ErrInvalidRide, got %v", err)
	}

	if err := tracker.Add(ride); err != nil {
		t.Fatalf("add ride failed: %v", err)
	}
	if err := tracker.Add(ride); !stderrors.Is(err, appErrors.ErrDuplicateRide) {
		t.Fatalf("expected ErrDuplicateRide, got %v", err)
	}

	if _, err := tracker.Get(""); !stderrors.Is(err, appErrors.ErrRideNotFound) {
		t.Fatalf("expected ErrRideNotFound for empty ID, got %v", err)
	}
	if _, err := tracker.Remove("UNKNOWN"); !stderrors.Is(err, appErrors.ErrRideNotFound) {
		t.Fatalf("expected ErrRideNotFound for unknown ID, got %v", err)
	}
}
