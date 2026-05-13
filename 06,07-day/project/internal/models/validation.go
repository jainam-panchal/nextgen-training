package models

import "strings"

func IsValidLocation(loc Location) bool {
	return loc.Lat >= -90 &&
		loc.Lat <= 90 &&
		loc.Lng >= -180 &&
		loc.Lng <= 180
}

func IsValidDriverStatus(status DriverStatus) bool {
	switch status {
	case DriverAvailable, DriverBusy, DriverOffline:
		return true
	default:
		return false
	}
}

func IsValidRideStatus(status RideStatus) bool {
	switch status {
	case RideRequested, RideAssigned, RideCompleted, RideCancelled:
		return true
	default:
		return false
	}
}

func IsValidDriver(d *Driver) bool {
	if d == nil {
		return false
	}

	if strings.TrimSpace(d.ID) == "" {
		return false
	}

	if strings.TrimSpace(d.Name) == "" {
		return false
	}

	if !IsValidLocation(d.Location) {
		return false
	}

	if !IsValidDriverStatus(d.Status) {
		return false
	}

	if d.Rating < 0 || d.Rating > 5 {
		return false
	}

	return true
}

func IsValidRider(r *Rider) bool {
	if r == nil {
		return false
	}

	if strings.TrimSpace(r.ID) == "" {
		return false
	}

	if strings.TrimSpace(r.Name) == "" {
		return false
	}

	if !IsValidLocation(r.Location) {
		return false
	}

	if strings.TrimSpace(r.PaymentMethod) == "" {
		return false
	}

	return true
}

func IsValidRideRequest(req *RideRequest) bool {
	if req == nil {
		return false
	}

	if strings.TrimSpace(req.ID) == "" {
		return false
	}

	if strings.TrimSpace(req.RiderID) == "" {
		return false
	}

	if !IsValidLocation(req.Pickup) {
		return false
	}

	if !IsValidLocation(req.Dropoff) {
		return false
	}

	if req.Pickup == req.Dropoff {
		return false
	}

	if req.RequestTime.IsZero() {
		return false
	}

	return true
}

func IsValidRide(ride *Ride) bool {
	if ride == nil {
		return false
	}

	if strings.TrimSpace(ride.ID) == "" {
		return false
	}

	if strings.TrimSpace(ride.RiderID) == "" {
		return false
	}

	if ride.Status != RideCancelled && strings.TrimSpace(ride.DriverID) == "" {
		return false
	}

	if !IsValidLocation(ride.Pickup) {
		return false
	}

	if !IsValidLocation(ride.Dropoff) {
		return false
	}

	if ride.Pickup == ride.Dropoff {
		return false
	}

	if !IsValidRideStatus(ride.Status) {
		return false
	}

	if ride.RequestTime.IsZero() {
		return false
	}

	if ride.Fare < 0 {
		return false
	}

	if ride.DistanceKm < 0 {
		return false
	}

	return true
}
