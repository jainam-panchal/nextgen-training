package rides

import "ride-sharing/internal/models"

type ActiveRideTracker interface {
	Add(ride *models.Ride) error
	Get(rideID string) (*models.Ride, error)
	Remove(rideID string) (*models.Ride, error)
	List() []*models.Ride
	Len() int
	IsEmpty() bool
}

type RideHistoryStore interface {
	Save(ride *models.Ride) error
	Get(rideID string) (*models.Ride, error)
	List() []*models.Ride
	ListByDriver(driverID string) []*models.Ride
}
