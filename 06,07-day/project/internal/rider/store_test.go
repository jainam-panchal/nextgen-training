package rider

import (
	stderrors "errors"
	"testing"

	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
)

func newValidRider() *models.Rider {
	return &models.Rider{
		ID:            models.NewRiderID(),
		Name:          "Jainam",
		Location:      models.Location{Lat: 19.08, Lng: 72.88},
		PaymentMethod: "UPI",
	}
}

// Tests register/get flow with duplicate and invalid rider checks.
func TestRegisterAndGet(t *testing.T) {
	store := NewMemoryStore()
	r := newValidRider()

	if err := store.Register(r); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if err := store.Register(r); !stderrors.Is(err, appErrors.ErrDuplicateRider) {
		t.Fatalf("expected ErrDuplicateRider, got %v", err)
	}
	if err := store.Register(&models.Rider{}); !stderrors.Is(err, appErrors.ErrInvalidRider) {
		t.Fatalf("expected ErrInvalidRider, got %v", err)
	}

	got, err := store.Get(r.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.ID != r.ID {
		t.Fatalf("expected rider %s, got %s", r.ID, got.ID)
	}

	if _, err := store.Get(""); !stderrors.Is(err, appErrors.ErrRiderNotFound) {
		t.Fatalf("expected ErrRiderNotFound for empty id, got %v", err)
	}
	if _, err := store.Get("UNKNOWN"); !stderrors.Is(err, appErrors.ErrRiderNotFound) {
		t.Fatalf("expected ErrRiderNotFound for unknown id, got %v", err)
	}
}

// Tests location update validation and success path.
func TestUpdateLocation(t *testing.T) {
	store := NewMemoryStore()
	r := newValidRider()
	_ = store.Register(r)

	if err := store.UpdateLocation("", models.Location{Lat: 19.1, Lng: 72.9}); !stderrors.Is(err, appErrors.ErrRiderNotFound) {
		t.Fatalf("expected ErrRiderNotFound, got %v", err)
	}
	if err := store.UpdateLocation(r.ID, models.Location{Lat: 200, Lng: 72.9}); !stderrors.Is(err, appErrors.ErrInvalidLocation) {
		t.Fatalf("expected ErrInvalidLocation, got %v", err)
	}
	if err := store.UpdateLocation("UNKNOWN", models.Location{Lat: 19.1, Lng: 72.9}); !stderrors.Is(err, appErrors.ErrRiderNotFound) {
		t.Fatalf("expected ErrRiderNotFound, got %v", err)
	}

	newLoc := models.Location{Lat: 19.1, Lng: 72.9}
	if err := store.UpdateLocation(r.ID, newLoc); err != nil {
		t.Fatalf("update location failed: %v", err)
	}

	got, _ := store.Get(r.ID)
	if got.Location != newLoc {
		t.Fatalf("expected location %+v, got %+v", newLoc, got.Location)
	}
}

// Tests ride history append and input validation.
func TestAddRideToHistory(t *testing.T) {
	store := NewMemoryStore()
	r := newValidRider()
	_ = store.Register(r)

	if err := store.AddRideToHistory(r.ID, "RID-000001"); err != nil {
		t.Fatalf("add ride history failed: %v", err)
	}
	got, _ := store.Get(r.ID)
	if len(got.RideHistory) != 1 || got.RideHistory[0] != "RID-000001" {
		t.Fatalf("expected ride history to contain RID-000001")
	}

	if err := store.AddRideToHistory("", "RID-000002"); !stderrors.Is(err, appErrors.ErrRiderNotFound) {
		t.Fatalf("expected ErrRiderNotFound, got %v", err)
	}
	if err := store.AddRideToHistory(r.ID, ""); !stderrors.Is(err, appErrors.ErrInvalidRide) {
		t.Fatalf("expected ErrInvalidRide, got %v", err)
	}
	if err := store.AddRideToHistory("UNKNOWN", "RID-000002"); !stderrors.Is(err, appErrors.ErrRiderNotFound) {
		t.Fatalf("expected ErrRiderNotFound, got %v", err)
	}
}
