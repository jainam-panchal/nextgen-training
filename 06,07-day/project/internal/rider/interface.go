package rider

import "ride-sharing/internal/models"

type RiderReader interface {
	Get(riderID string) (*models.Rider, error)
}

type RiderManager interface {
	RiderReader

	Register(rider *models.Rider) error
	UpdateLocation(riderID string, loc models.Location) error
}

type RiderHistoryStore interface {
	RiderReader

	AddRideToHistory(riderID string, rideID string) error
}

type Store interface {
	RiderManager
	RiderHistoryStore
}
