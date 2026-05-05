// Package tracker contains the stock tracking domain model and rolling window logic.
package tracker

import "time"

const (
	InitialWindowCapacity = 10
)

// StockStatus represents the freshness state of a tracker.
type StockStatus string

const (
	StatusNoData StockStatus = "NO_DATA"
	StatusLive   StockStatus = "LIVE"
	StatusStale  StockStatus = "STALE"
)

type PricePoint struct {
	Price     float64
	Volume    int
	Timestamp time.Time
}

type StockTracker struct {
	Symbol             string
	Status             StockStatus
	Window             []PricePoint
	WindowSum          float64
	WindowSize         int
	LastGeneratedPrice float64
}
