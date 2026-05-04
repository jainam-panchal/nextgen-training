package garage

import "errors"

var (
	ErrVehicleAlreadyParked = errors.New("vehicle already parked")
	ErrVehicleNotParked     = errors.New("vehicle not parked")
	ErrVehicleNotFound      = errors.New("vehicle not found")
	ErrFloorFull            = errors.New("floor full")
	ErrInvalidFloor         = errors.New("invalid floor")
	ErrInvalidPlate         = errors.New("invalid plate")
	ErrPlateRequired        = errors.New("plate required")
)
