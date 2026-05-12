package dispatch

import (
	"sort"
	"sync"
	"time"

	"ride-sharing/internal/config"
	"ride-sharing/internal/driver"
	appErrors "ride-sharing/internal/errors"
	"ride-sharing/internal/geo"
	"ride-sharing/internal/models"
	"ride-sharing/internal/ridequeue"
	"ride-sharing/internal/rider"
	"ride-sharing/internal/rides"
)

type Dispatcher struct {
	mu sync.Mutex

	cfg config.Config

	driverStore driver.DispatchDriverStore
	riderStore  rider.Store

	requestQueue ridequeue.RideRequestQueue
	activeRides  rides.ActiveRideTracker
	rideHistory  rides.RideHistoryStore
}

func NewDispatcher(
	cfg config.Config,
	driverStore driver.DispatchDriverStore,
	riderStore rider.Store,
	requestQueue ridequeue.RideRequestQueue,
	activeRides rides.ActiveRideTracker,
	rideHistory rides.RideHistoryStore,
) *Dispatcher {
	return &Dispatcher{
		cfg:          cfg,
		driverStore:  driverStore,
		riderStore:   riderStore,
		requestQueue: requestQueue,
		activeRides:  activeRides,
		rideHistory:  rideHistory,
	}
}

func (d *Dispatcher) RequestRide(
	riderID string,
	pickup models.Location,
	dropoff models.Location,
	requestTime time.Time,
) (*models.RideRequest, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !models.IsValidLocation(pickup) || !models.IsValidLocation(dropoff) {
		return nil, appErrors.ErrInvalidLocation
	}

	if pickup == dropoff {
		return nil, appErrors.ErrInvalidRideRequest
	}

	if requestTime.After(time.Now()) {
		return nil, appErrors.ErrInvalidRideRequest
	}

	_, err := d.riderStore.Get(riderID)
	if err != nil {
		return nil, err
	}

	request := &models.RideRequest{
		ID:          models.NewRequestID(),
		RiderID:     riderID,
		Pickup:      pickup,
		Dropoff:     dropoff,
		RequestTime: requestTime,
	}

	if err := d.requestQueue.Push(request); err != nil {
		return nil, err
	}

	return request, nil
}

func (d *Dispatcher) ProcessNextRequest(now time.Time) (*models.Ride, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	request, err := d.requestQueue.Peek()
	if err != nil {
		return nil, err
	}

	if d.isRequestExpired(request, now) {
		expiredRequest, popErr := d.requestQueue.Pop()
		if popErr != nil {
			return nil, popErr
		}

		cancelledRide := &models.Ride{
			ID:          models.NewRideID(),
			RiderID:     expiredRequest.RiderID,
			DriverID:    "",
			Pickup:      expiredRequest.Pickup,
			Dropoff:     expiredRequest.Dropoff,
			Status:      models.RideCancelled,
			RequestTime: expiredRequest.RequestTime,
			EndTime:     now,
			Fare:        0,
			DistanceKm:  0,
		}

		if err := d.rideHistory.Save(cancelledRide); err != nil {
			return nil, err
		}

		return nil, appErrors.ErrRideRequestExpired
	}

	nearestDriver, pickupDistanceKm, err := d.driverStore.FindNearestAvailable(
		request.Pickup,
		d.cfg.MaxPickupDistanceKm,
	)
	if err != nil {
		return nil, appErrors.ErrNoDriverAvailable
	}

	if err := d.driverStore.Assign(nearestDriver.ID); err != nil {
		return nil, err
	}

	_, err = d.requestQueue.Pop()
	if err != nil {
		return nil, err
	}

	tripDistanceKm := geo.DistanceKm(request.Pickup, request.Dropoff)
	fare := d.calculateFare(tripDistanceKm)

	ride := &models.Ride{
		ID:          models.NewRideID(),
		RiderID:     request.RiderID,
		DriverID:    nearestDriver.ID,
		Pickup:      request.Pickup,
		Dropoff:     request.Dropoff,
		Status:      models.RideAssigned,
		RequestTime: request.RequestTime,
		StartTime:   now,
		Fare:        fare,
		DistanceKm:  tripDistanceKm,
	}

	if err := d.activeRides.Add(ride); err != nil {
		_ = d.driverStore.Release(nearestDriver.ID)
		return nil, err
	}

	_ = pickupDistanceKm

	return ride, nil
}

func (d *Dispatcher) CompleteRide(rideID string, completedAt time.Time) (*models.Ride, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	ride, err := d.activeRides.Remove(rideID)
	if err != nil {
		return nil, err
	}

	originalStatus := ride.Status
	originalEndTime := ride.EndTime

	ride.Status = models.RideCompleted
	ride.EndTime = completedAt

	if err := d.rideHistory.Save(ride); err != nil {
		ride.Status = originalStatus
		ride.EndTime = originalEndTime
		_ = d.activeRides.Add(ride)
		return nil, err
	}

	if err := d.driverStore.UpdateLocation(ride.DriverID, ride.Dropoff); err != nil {
		_ = d.driverStore.Release(ride.DriverID)
		return nil, err
	}

	if err := d.driverStore.AddRideToHistory(ride.DriverID, ride.ID); err != nil {
		_ = d.driverStore.Release(ride.DriverID)
		return nil, err
	}

	if err := d.riderStore.AddRideToHistory(ride.RiderID, ride.ID); err != nil {
		_ = d.driverStore.Release(ride.DriverID)
		return nil, err
	}

	if err := d.driverStore.Release(ride.DriverID); err != nil {
		return nil, err
	}

	return ride, nil
}

func (d *Dispatcher) FindNNearestDrivers(
	loc models.Location,
	limit int,
) ([]*models.Driver, error) {
	return d.driverStore.FindNNearestAvailable(
		loc,
		limit,
		d.cfg.MaxPickupDistanceKm,
	)
}

func (d *Dispatcher) DriverEarningsToday(driverID string, now time.Time) (float64, error) {
	_, err := d.driverStore.Get(driverID)
	if err != nil {
		return 0, err
	}

	startOfDay := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)

	return d.driverEarningsSince(driverID, startOfDay), nil
}

func (d *Dispatcher) DriverEarningsThisWeek(driverID string, now time.Time) (float64, error) {
	_, err := d.driverStore.Get(driverID)
	if err != nil {
		return 0, err
	}

	weekdayOffset := int(now.Weekday())
	if weekdayOffset == 0 {
		weekdayOffset = 6
	} else {
		weekdayOffset--
	}

	startOfWeekDate := now.AddDate(0, 0, -weekdayOffset)

	startOfWeek := time.Date(
		startOfWeekDate.Year(),
		startOfWeekDate.Month(),
		startOfWeekDate.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)

	return d.driverEarningsSince(driverID, startOfWeek), nil
}

func (d *Dispatcher) AverageWaitTime() time.Duration {
	completedRides := d.rideHistory.List()

	var totalWait time.Duration
	completedRideCount := 0

	for _, ride := range completedRides {
		if ride.Status != models.RideCompleted {
			continue
		}

		if ride.RequestTime.IsZero() || ride.StartTime.IsZero() {
			continue
		}

		wait := ride.StartTime.Sub(ride.RequestTime)
		if wait < 0 {
			continue
		}

		totalWait += wait
		completedRideCount++
	}

	if completedRideCount == 0 {
		return 0
	}

	return totalWait / time.Duration(completedRideCount)
}

func (d *Dispatcher) calculateFare(distanceKm float64) float64 {
	return d.cfg.BaseFare + d.cfg.FarePerKm*distanceKm
}

func (d *Dispatcher) isRequestExpired(
	request *models.RideRequest,
	now time.Time,
) bool {
	return now.Sub(request.RequestTime) > d.cfg.RequestTimeout
}

func (d *Dispatcher) driverEarningsSince(
	driverID string,
	startTime time.Time,
) float64 {
	driverRides := d.rideHistory.ListByDriver(driverID)

	var earnings float64

	for _, ride := range driverRides {
		if ride.Status != models.RideCompleted {
			continue
		}

		if ride.EndTime.Before(startTime) {
			continue
		}

		earnings += ride.Fare
	}

	return earnings
}

func (d *Dispatcher) BusiestZones(limit int) []models.ZoneRequestCount {
	if limit <= 0 {
		return []models.ZoneRequestCount{}
	}

	requests := d.requestQueue.Snapshot()
	if len(requests) == 0 {
		return []models.ZoneRequestCount{}
	}

	countByBlock := make(map[models.BlockID]int)
	for _, request := range requests {
		if request == nil {
			continue
		}

		block := geo.GetBlockID(request.Pickup, d.cfg.CellSizeDegrees)
		countByBlock[block]++
	}

	zoneCounts := make([]models.ZoneRequestCount, 0, len(countByBlock))
	for block, count := range countByBlock {
		zoneCounts = append(zoneCounts, models.ZoneRequestCount{
			BlockID: block,
			Count:   count,
		})
	}

	// Sort by highest request count first; use block coordinates
	sort.Slice(zoneCounts, func(i, j int) bool {
		if zoneCounts[i].Count == zoneCounts[j].Count {
			left := zoneCounts[i].BlockID
			right := zoneCounts[j].BlockID
			return left.LatBucket < right.LatBucket ||
				(left.LatBucket == right.LatBucket && left.LngBucket < right.LngBucket)
		}

		return zoneCounts[i].Count > zoneCounts[j].Count
	})

	if limit > len(zoneCounts) {
		limit = len(zoneCounts)
	}

	return zoneCounts[:limit]
}
