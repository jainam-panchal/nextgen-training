package models

type DriverStatus string

const (
	DriverAvailable DriverStatus = "available"
	DriverBusy      DriverStatus = "busy"
	DriverOffline   DriverStatus = "offline"
)

type RideStatus string

const (
	RideRequested RideStatus = "requested"
	RideAssigned  RideStatus = "assigned"
	RideCompleted RideStatus = "completed"
	RideCancelled RideStatus = "cancelled"
)
