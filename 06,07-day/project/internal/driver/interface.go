package driver

import "ride-sharing/internal/models"

type DriverReader interface {
	Get(driverID string) (*models.Driver, error)
	FindNearestAvailable(
		loc models.Location,
		maxDistanceKm float64,
	) (*models.Driver, float64, error)
	FindNNearestAvailable(
		loc models.Location,
		limit int,
		maxDistanceKm float64,
	) ([]*models.Driver, error)
}

type DriverManager interface {
	DriverReader

	Register(driver *models.Driver) error
	GoOnline(driverID string) error
	GoOffline(driverID string) error
	UpdateLocation(driverID string, loc models.Location) error
}

type DispatchDriverStore interface {
	DriverReader

	Assign(driverID string) error
	Release(driverID string) error
	UpdateLocation(driverID string, loc models.Location) error
	AddRideToHistory(driverID string, rideID string) error
}

type Store interface {
	DriverManager
	DispatchDriverStore
}
