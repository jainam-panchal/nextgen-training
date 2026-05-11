package models

import (
	"fmt"
	"sync/atomic"
)

var driverCounter uint64
var riderCounter uint64
var rideCounter uint64
var requestCounter uint64

func NewDriverID() string {
	id := atomic.AddUint64(&driverCounter, 1)
	return fmt.Sprintf("DRV-%06d", id)
}

func NewRiderID() string {
	id := atomic.AddUint64(&riderCounter, 1)
	return fmt.Sprintf("RDR-%06d", id)
}

func NewRideID() string {
	id := atomic.AddUint64(&rideCounter, 1)
	return fmt.Sprintf("RID-%06d", id)
}

func NewRequestID() string {
	id := atomic.AddUint64(&requestCounter, 1)
	return fmt.Sprintf("REQ-%06d", id)
}
