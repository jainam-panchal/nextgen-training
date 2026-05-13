package geo

import (
	"math"

	"ride-sharing/internal/models"
)

func GetBlockID(loc models.Location, cellSizeDegrees float64) models.BlockID {
	return models.BlockID{
		LatBucket: int(math.Floor(loc.Lat / cellSizeDegrees)),
		LngBucket: int(math.Floor(loc.Lng / cellSizeDegrees)),
	}
}

func NeighborBlocks(center models.BlockID, ring int) []models.BlockID {
	size := (2*ring + 1) * (2*ring + 1)
	blocks := make([]models.BlockID, 0, size)

	for latOffset := -ring; latOffset <= ring; latOffset++ {
		for lngOffset := -ring; lngOffset <= ring; lngOffset++ {
			blocks = append(blocks, models.BlockID{
				LatBucket: center.LatBucket + latOffset,
				LngBucket: center.LngBucket + lngOffset,
			})
		}
	}

	return blocks
}

func DistanceKm(a, b models.Location) float64 {
	dx := (a.Lat - b.Lat) * 111.0
	dy := (a.Lng - b.Lng) * 111.0 * math.Cos(a.Lat*math.Pi/180)

	return math.Sqrt(dx*dx + dy*dy)
}
