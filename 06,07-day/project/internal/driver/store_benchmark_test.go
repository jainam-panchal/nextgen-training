package driver

import (
	"fmt"
	"testing"

	"ride-sharing/internal/config"
	"ride-sharing/internal/models"
)

func BenchmarkFindNearestAvailable(b *testing.B) {
	sizes := []int{100, 1_000, 10_000}
	query := models.Location{Lat: 19.08, Lng: 72.88}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			store := NewMemoryStore(config.DefaultConfig())

			for i := 0; i < n; i++ {
				lat := 19.00 + float64(i%1000)*0.00001
				lng := 72.80 + float64((i/1000)%1000)*0.00001
				d := &models.Driver{
					ID:       models.NewDriverID(),
					Name:     "Driver",
					Location: models.Location{Lat: lat, Lng: lng},
					Status:   models.DriverOffline,
					Rating:   4.5,
				}
				_ = store.Register(d)
				_ = store.GoOnline(d.ID)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = store.FindNearestAvailable(query, 5)
			}
		})
	}
}
