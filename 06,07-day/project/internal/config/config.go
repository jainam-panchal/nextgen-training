package config

import (
	"math"
	"time"
)

type Config struct {
	CellSizeDegrees     float64
	MaxPickupDistanceKm float64
	RequestTimeout      time.Duration
	BaseFare            float64
	FarePerKm           float64
}

func DefaultConfig() Config {
	return Config{
		CellSizeDegrees:     0.01,
		MaxPickupDistanceKm: 5.0,
		RequestTimeout:      10 * time.Minute,
		BaseFare:            50.0,
		FarePerKm:           12.0,
	}
}

func (c Config) NeighborRing() int {
	cellSizeKm := c.CellSizeDegrees * 111.0
	return int(math.Ceil(c.MaxPickupDistanceKm / cellSizeKm))
}
