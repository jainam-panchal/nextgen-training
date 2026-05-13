package rides

import (
	"fmt"
	"testing"
	"time"

	"ride-sharing/internal/models"
)

func benchmarkRide() *models.Ride {
	now := time.Now()
	return &models.Ride{
		ID:          models.NewRideID(),
		RiderID:     models.NewRiderID(),
		DriverID:    models.NewDriverID(),
		Pickup:      models.Location{Lat: 19.08, Lng: 72.88},
		Dropoff:     models.Location{Lat: 19.09, Lng: 72.89},
		Status:      models.RideAssigned,
		RequestTime: now.Add(-2 * time.Minute),
		StartTime:   now.Add(-1 * time.Minute),
		Fare:        120,
		DistanceKm:  5,
	}
}

func BenchmarkActiveRideTrackerAddRemove(b *testing.B) {
	sizes := []int{100, 1_000, 10_000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tracker := NewMemoryActiveRideTracker()
			ids := make([]string, 0, n)

			for i := 0; i < n; i++ {
				r := benchmarkRide()
				_ = tracker.Add(r)
				ids = append(ids, r.ID)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := benchmarkRide()
				_ = tracker.Add(r)
				_, _ = tracker.Remove(r.ID)
			}

			_ = ids
		})
	}
}
