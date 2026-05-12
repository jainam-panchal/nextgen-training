package models

import (
	"testing"
	"time"
)

// Tests location boundary validation.
func TestIsValidLocation(t *testing.T) {
	if !IsValidLocation(Location{Lat: 19.08, Lng: 72.88}) {
		t.Fatalf("expected valid location")
	}
	if IsValidLocation(Location{Lat: -91, Lng: 72.88}) {
		t.Fatalf("expected invalid latitude")
	}
	if IsValidLocation(Location{Lat: 19.08, Lng: 181}) {
		t.Fatalf("expected invalid longitude")
	}
}

// Tests driver/rider status validators.
func TestStatusValidators(t *testing.T) {
	if !IsValidDriverStatus(DriverAvailable) || !IsValidDriverStatus(DriverBusy) || !IsValidDriverStatus(DriverOffline) {
		t.Fatalf("expected all driver statuses to be valid")
	}
	if IsValidDriverStatus(DriverStatus("x")) {
		t.Fatalf("expected unknown driver status to be invalid")
	}

	if !IsValidRideStatus(RideRequested) || !IsValidRideStatus(RideAssigned) || !IsValidRideStatus(RideCompleted) || !IsValidRideStatus(RideCancelled) {
		t.Fatalf("expected all ride statuses to be valid")
	}
	if IsValidRideStatus(RideStatus("x")) {
		t.Fatalf("expected unknown ride status to be invalid")
	}
}

// Tests driver and rider object validation.
func TestEntityValidators(t *testing.T) {
	validDriver := &Driver{
		ID:       NewDriverID(),
		Name:     "Amit",
		Location: Location{Lat: 19.08, Lng: 72.88},
		Status:   DriverOffline,
		Rating:   4.8,
	}
	if !IsValidDriver(validDriver) {
		t.Fatalf("expected valid driver")
	}
	invalidDriver := *validDriver
	invalidDriver.Rating = 6
	if IsValidDriver(&invalidDriver) {
		t.Fatalf("expected invalid driver rating")
	}

	validRider := &Rider{
		ID:            NewRiderID(),
		Name:          "Jainam",
		Location:      Location{Lat: 19.08, Lng: 72.88},
		PaymentMethod: "UPI",
	}
	if !IsValidRider(validRider) {
		t.Fatalf("expected valid rider")
	}
	invalidRider := *validRider
	invalidRider.PaymentMethod = ""
	if IsValidRider(&invalidRider) {
		t.Fatalf("expected invalid rider payment")
	}
}

// Tests ride request validator rules.
func TestIsValidRideRequest(t *testing.T) {
	now := time.Now()
	req := &RideRequest{
		ID:          NewRequestID(),
		RiderID:     NewRiderID(),
		Pickup:      Location{Lat: 19.08, Lng: 72.88},
		Dropoff:     Location{Lat: 19.09, Lng: 72.89},
		RequestTime: now,
	}
	if !IsValidRideRequest(req) {
		t.Fatalf("expected valid ride request")
	}

	bad := *req
	bad.Dropoff = bad.Pickup
	if IsValidRideRequest(&bad) {
		t.Fatalf("expected invalid request for same pickup/dropoff")
	}
}

// Tests ride validator including cancelled ride with empty driver id.
func TestIsValidRide(t *testing.T) {
	now := time.Now()
	ride := &Ride{
		ID:          NewRideID(),
		RiderID:     NewRiderID(),
		DriverID:    NewDriverID(),
		Pickup:      Location{Lat: 19.08, Lng: 72.88},
		Dropoff:     Location{Lat: 19.09, Lng: 72.89},
		Status:      RideCompleted,
		RequestTime: now.Add(-2 * time.Minute),
		StartTime:   now.Add(-time.Minute),
		EndTime:     now,
		Fare:        100,
		DistanceKm:  4.5,
	}
	if !IsValidRide(ride) {
		t.Fatalf("expected valid completed ride")
	}

	cancelled := *ride
	cancelled.Status = RideCancelled
	cancelled.DriverID = ""
	if !IsValidRide(&cancelled) {
		t.Fatalf("expected cancelled ride with empty driver id to be valid")
	}

	invalid := *ride
	invalid.DriverID = ""
	if IsValidRide(&invalid) {
		t.Fatalf("expected non-cancelled ride with empty driver id to be invalid")
	}
}
