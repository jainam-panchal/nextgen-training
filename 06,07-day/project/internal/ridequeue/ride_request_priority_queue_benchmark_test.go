package ridequeue

import (
	"fmt"
	"testing"
	"time"

	"ride-sharing/internal/models"
)

func benchmarkRequest(at time.Time) *models.RideRequest {
	return &models.RideRequest{
		ID:          models.NewRequestID(),
		RiderID:     models.NewRiderID(),
		Pickup:      models.Location{Lat: 19.08, Lng: 72.88},
		Dropoff:     models.Location{Lat: 19.09, Lng: 72.89},
		RequestTime: at,
	}
}

func BenchmarkRideRequestQueuePushPop(b *testing.B) {
	sizes := []int{100, 1_000, 10_000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			q := NewRideRequestPriorityQueue()
			base := time.Now()

			for i := 0; i < n; i++ {
				_ = q.Push(benchmarkRequest(base.Add(time.Duration(i) * time.Millisecond)))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = q.Push(benchmarkRequest(base.Add(time.Duration(n+i) * time.Millisecond)))
				_, _ = q.Pop()
			}
		})
	}
}
