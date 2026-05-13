package ridequeue

import "ride-sharing/internal/models"

type RideRequestQueue interface {
	Push(request *models.RideRequest) error
	Peek() (*models.RideRequest, error)
	Pop() (*models.RideRequest, error)
	Snapshot() []*models.RideRequest
	Len() int
	IsEmpty() bool
}
