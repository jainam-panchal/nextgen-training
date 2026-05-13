package dispatch

import (
	"fmt"
	"testing"
	"time"

	"ride-sharing/internal/config"
	"ride-sharing/internal/driver"
	"ride-sharing/internal/models"
	"ride-sharing/internal/ridequeue"
	"ride-sharing/internal/rider"
	"ride-sharing/internal/rides"
)

func benchmarkDispatcherSetup(driverCount int) (*Dispatcher, *models.Rider) {
	cfg := config.DefaultConfig()
	driverStore := driver.NewMemoryStore(cfg)
	riderStore := rider.NewMemoryStore()
	requestQueue := ridequeue.NewRideRequestPriorityQueue()
	activeRideTracker := rides.NewMemoryActiveRideTracker()
	rideHistoryStore := rides.NewMemoryRideHistoryStore()

	d := NewDispatcher(cfg, driverStore, riderStore, requestQueue, activeRideTracker, rideHistoryStore)

	r := &models.Rider{
		ID:            models.NewRiderID(),
		Name:          "Rider",
		Location:      models.Location{Lat: 19.08, Lng: 72.88},
		PaymentMethod: "UPI",
	}
	_ = riderStore.Register(r)

	for i := 0; i < driverCount; i++ {
		drv := &models.Driver{
			ID:       models.NewDriverID(),
			Name:     "Driver",
			Location: models.Location{Lat: 19.08 + float64(i)*0.00001, Lng: 72.88 + float64(i)*0.00001},
			Status:   models.DriverOffline,
			Rating:   4.5,
		}
		_ = driverStore.Register(drv)
		_ = driverStore.GoOnline(drv.ID)
	}

	return d, r
}

func BenchmarkDispatcherRequestAndProcess(b *testing.B) {
	sizes := []int{10, 100, 1_000}
	pickup := models.Location{Lat: 19.08, Lng: 72.88}
	dropoff := models.Location{Lat: 19.09, Lng: 72.89}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("drivers=%d", n), func(b *testing.B) {
			d, r := benchmarkDispatcherSetup(n)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tm := time.Now().Add(-time.Second)
				_, _ = d.RequestRide(r.ID, pickup, dropoff, tm)
				ride, _ := d.ProcessNextRequest(time.Now())
				if ride != nil {
					_, _ = d.CompleteRide(ride.ID, time.Now())
				}
			}
		})
	}
}
