package dispatch

import (
	stderrors "errors"
	"testing"
	"time"

	"ride-sharing/internal/config"
	"ride-sharing/internal/driver"
	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/models"
	"ride-sharing/internal/ridequeue"
	"ride-sharing/internal/rider"
	"ride-sharing/internal/rides"
)

func setupDispatcher(t *testing.T) (
	*Dispatcher,
	*driver.MemoryStore,
	*rider.MemoryStore,
	*ridequeue.RideRequestPriorityQueue,
	*rides.MemoryActiveRideTracker,
	*rides.MemoryRideHistoryStore,
) {
	t.Helper()

	cfg := config.DefaultConfig()
	driverStore := driver.NewMemoryStore(cfg)
	riderStore := rider.NewMemoryStore()
	requestQueue := ridequeue.NewRideRequestPriorityQueue()
	activeRideTracker := rides.NewMemoryActiveRideTracker()
	rideHistoryStore := rides.NewMemoryRideHistoryStore()

	dispatcher := NewDispatcher(
		cfg,
		driverStore,
		riderStore,
		requestQueue,
		activeRideTracker,
		rideHistoryStore,
	)

	return dispatcher, driverStore, riderStore, requestQueue, activeRideTracker, rideHistoryStore
}

// mustRegisterRider creates and stores a valid rider for test setup.
func mustRegisterRider(t *testing.T, riderStore *rider.MemoryStore, loc models.Location) *models.Rider {
	t.Helper()

	newRider := &models.Rider{
		ID:            models.NewRiderID(),
		Name:          "Rider",
		Location:      loc,
		PaymentMethod: "UPI",
	}

	if err := riderStore.Register(newRider); err != nil {
		t.Fatalf("register rider: %v", err)
	}

	return newRider
}

// mustRegisterOnlineDriver creates a driver and moves it to available state.
func mustRegisterOnlineDriver(t *testing.T, driverStore *driver.MemoryStore, loc models.Location) *models.Driver {
	t.Helper()

	newDriver := &models.Driver{
		ID:       models.NewDriverID(),
		Name:     "Driver",
		Location: loc,
		Status:   models.DriverOffline,
		Rating:   4.8,
	}

	if err := driverStore.Register(newDriver); err != nil {
		t.Fatalf("register driver: %v", err)
	}

	if err := driverStore.GoOnline(newDriver.ID); err != nil {
		t.Fatalf("go online: %v", err)
	}

	return newDriver
}

// Tests request validation: invalid locations, same pickup/dropoff, unknown rider, and success.
func TestRequestRideValidation(t *testing.T) {
	dispatcher, _, riderStore, _, _, _ := setupDispatcher(t)
	validRider := mustRegisterRider(t, riderStore, models.Location{Lat: 19.0800, Lng: 72.8800})

	now := time.Now()

	_, err := dispatcher.RequestRide(validRider.ID, models.Location{Lat: 200, Lng: 72.88}, models.Location{Lat: 19.09, Lng: 72.89}, now)
	if !stderrors.Is(err, appErrors.ErrInvalidLocation) {
		t.Fatalf("expected ErrInvalidLocation, got %v", err)
	}

	_, err = dispatcher.RequestRide(validRider.ID, models.Location{Lat: 19.08, Lng: 72.88}, models.Location{Lat: 19.08, Lng: 72.88}, now)
	if !stderrors.Is(err, appErrors.ErrInvalidRideRequest) {
		t.Fatalf("expected ErrInvalidRideRequest, got %v", err)
	}

	_, err = dispatcher.RequestRide("UNKNOWN", models.Location{Lat: 19.08, Lng: 72.88}, models.Location{Lat: 19.09, Lng: 72.89}, now)
	if !stderrors.Is(err, appErrors.ErrRiderNotFound) {
		t.Fatalf("expected ErrRiderNotFound, got %v", err)
	}

	request, err := dispatcher.RequestRide(validRider.ID, models.Location{Lat: 19.08, Lng: 72.88}, models.Location{Lat: 19.09, Lng: 72.89}, now)
	if err != nil {
		t.Fatalf("valid request failed: %v", err)
	}
	if request.ID == "" {
		t.Fatalf("expected request ID to be set")
	}
}

// Tests dispatch behavior when no driver is available: request should remain queued.
func TestProcessNextRequestNoDriverKeepsQueue(t *testing.T) {
	dispatcher, _, riderStore, requestQueue, _, _ := setupDispatcher(t)
	newRider := mustRegisterRider(t, riderStore, models.Location{Lat: 19.0800, Lng: 72.8800})

	_, err := dispatcher.RequestRide(
		newRider.ID,
		models.Location{Lat: 19.0800, Lng: 72.8800},
		models.Location{Lat: 19.0900, Lng: 72.8900},
		time.Now(),
	)
	if err != nil {
		t.Fatalf("request ride: %v", err)
	}

	_, err = dispatcher.ProcessNextRequest(time.Now())
	if !stderrors.Is(err, appErrors.ErrNoDriverAvailable) {
		t.Fatalf("expected ErrNoDriverAvailable, got %v", err)
	}

	if got := requestQueue.Len(); got != 1 {
		t.Fatalf("expected queue len 1, got %d", got)
	}
}

// Tests expiry path: old request is cancelled, removed from queue, and saved to history.
func TestProcessNextRequestExpiredRequestCancelled(t *testing.T) {
	dispatcher, _, riderStore, requestQueue, _, rideHistoryStore := setupDispatcher(t)
	newRider := mustRegisterRider(t, riderStore, models.Location{Lat: 19.0800, Lng: 72.8800})

	requestTime := time.Now().Add(-11 * time.Minute)
	_, err := dispatcher.RequestRide(
		newRider.ID,
		models.Location{Lat: 19.0800, Lng: 72.8800},
		models.Location{Lat: 19.0900, Lng: 72.8900},
		requestTime,
	)
	if err != nil {
		t.Fatalf("request ride: %v", err)
	}

	_, err = dispatcher.ProcessNextRequest(time.Now())
	if !stderrors.Is(err, appErrors.ErrRideRequestExpired) {
		t.Fatalf("expected ErrRideRequestExpired, got %v", err)
	}

	if got := requestQueue.Len(); got != 0 {
		t.Fatalf("expected queue len 0, got %d", got)
	}

	history := rideHistoryStore.List()
	if len(history) != 1 {
		t.Fatalf("expected 1 history item, got %d", len(history))
	}
	if history[0].Status != models.RideCancelled {
		t.Fatalf("expected cancelled ride status, got %s", history[0].Status)
	}
	if history[0].RiderID != newRider.ID {
		t.Fatalf("expected rider ID %s, got %s", newRider.ID, history[0].RiderID)
	}
}

// Tests successful match flow: nearest driver assigned and ride becomes active.
func TestProcessNextRequestAssignsDriver(t *testing.T) {
	dispatcher, driverStore, riderStore, _, activeRideTracker, _ := setupDispatcher(t)
	newRider := mustRegisterRider(t, riderStore, models.Location{Lat: 19.0800, Lng: 72.8800})
	assignedDriver := mustRegisterOnlineDriver(t, driverStore, models.Location{Lat: 19.0810, Lng: 72.8810})

	requestTime := time.Now()
	_, err := dispatcher.RequestRide(
		newRider.ID,
		models.Location{Lat: 19.0800, Lng: 72.8800},
		models.Location{Lat: 19.0900, Lng: 72.8900},
		requestTime,
	)
	if err != nil {
		t.Fatalf("request ride: %v", err)
	}

	ride, err := dispatcher.ProcessNextRequest(time.Now())
	if err != nil {
		t.Fatalf("process next request: %v", err)
	}

	if ride.Status != models.RideAssigned {
		t.Fatalf("expected RideAssigned, got %s", ride.Status)
	}
	if ride.DriverID != assignedDriver.ID {
		t.Fatalf("expected driver %s, got %s", assignedDriver.ID, ride.DriverID)
	}
	if ride.Fare <= 0 {
		t.Fatalf("expected fare > 0, got %f", ride.Fare)
	}
	if got := activeRideTracker.Len(); got != 1 {
		t.Fatalf("expected active rides 1, got %d", got)
	}
}

// Tests completion flow: active ride moved to history and driver/rider histories updated.
func TestCompleteRideSuccess(t *testing.T) {
	dispatcher, driverStore, riderStore, _, activeRideTracker, _ := setupDispatcher(t)
	newRider := mustRegisterRider(t, riderStore, models.Location{Lat: 19.0800, Lng: 72.8800})
	assignedDriver := mustRegisterOnlineDriver(t, driverStore, models.Location{Lat: 19.0810, Lng: 72.8810})

	_, err := dispatcher.RequestRide(
		newRider.ID,
		models.Location{Lat: 19.0800, Lng: 72.8800},
		models.Location{Lat: 19.0900, Lng: 72.8900},
		time.Now(),
	)
	if err != nil {
		t.Fatalf("request ride: %v", err)
	}

	assignedRide, err := dispatcher.ProcessNextRequest(time.Now())
	if err != nil {
		t.Fatalf("process next request: %v", err)
	}

	completedAt := time.Now().Add(10 * time.Minute)
	completedRide, err := dispatcher.CompleteRide(assignedRide.ID, completedAt)
	if err != nil {
		t.Fatalf("complete ride: %v", err)
	}

	if completedRide.Status != models.RideCompleted {
		t.Fatalf("expected RideCompleted, got %s", completedRide.Status)
	}
	if completedRide.EndTime.IsZero() {
		t.Fatalf("expected end time to be set")
	}
	if got := activeRideTracker.Len(); got != 0 {
		t.Fatalf("expected active rides 0, got %d", got)
	}

	updatedDriver, err := driverStore.Get(assignedDriver.ID)
	if err != nil {
		t.Fatalf("get driver: %v", err)
	}
	if updatedDriver.Status != models.DriverAvailable {
		t.Fatalf("expected driver available, got %s", updatedDriver.Status)
	}
	if len(updatedDriver.RideHistory) != 1 || updatedDriver.RideHistory[0] != completedRide.ID {
		t.Fatalf("expected driver history to contain ride ID %s", completedRide.ID)
	}

	updatedRider, err := riderStore.Get(newRider.ID)
	if err != nil {
		t.Fatalf("get rider: %v", err)
	}
	if len(updatedRider.RideHistory) != 1 || updatedRider.RideHistory[0] != completedRide.ID {
		t.Fatalf("expected rider history to contain ride ID %s", completedRide.ID)
	}
}
