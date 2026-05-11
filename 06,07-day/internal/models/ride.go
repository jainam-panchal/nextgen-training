package models

import "time"

type RideRequest struct {
	ID          string
	RiderID     string
	Pickup      Location
	Dropoff     Location
	RequestTime time.Time
}

type Ride struct {
	ID          string
	RiderID     string
	DriverID    string
	Pickup      Location
	Dropoff     Location
	Status      RideStatus
	RequestTime time.Time
	StartTime   time.Time
	EndTime     time.Time
	Fare        float64
	DistanceKm  float64
}
