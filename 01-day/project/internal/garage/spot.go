// Package garage provides types and configuration for a parking garage.
package garage

import "time"

type Config struct {
	TotalSpots    int
	TotalFloors   int
	SpotsPerFloor int
	HourlyRate    int
}

type ParkingSpot struct {
	ID        int
	Floor     int
	Position  int
	Occupied  bool
	Plate     string
	EntryTime time.Time
}
