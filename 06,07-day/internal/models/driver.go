package models

type Driver struct {
	ID          string
	Name        string
	Location    Location
	Status      DriverStatus
	Rating      float64
	BlockID     BlockID
	RideHistory []string
}
