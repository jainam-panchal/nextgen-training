package tracker

import (
	"fmt"
	"math/rand"
	"time"
)

func NewStockTracker(symbol string, windowSize int) (*StockTracker, error) {
	if windowSize <= 0 {
		return nil, fmt.Errorf("window size must be greater than 0")
	}

	return &StockTracker{
		Symbol:             symbol,
		Status:             StatusNoData,
		Window:             make([]PricePoint, 0, InitialWindowCapacity),
		WindowSum:          0,
		WindowSize:         windowSize,
		LastGeneratedPrice: initialPriceForSymbol(symbol),
	}, nil
}

func (s *StockTracker) AddPrice(newPoint PricePoint) (oldCap int, newCap int) {
	oldCap = cap(s.Window)
	s.Window = append(s.Window, newPoint)
	newCap = cap(s.Window)
	s.Status = StatusLive

	s.WindowSum += newPoint.Price

	// slide the window
	if len(s.Window) > s.WindowSize {
		s.WindowSum -= s.Window[0].Price
		s.Window = s.Window[1:]

		// manual shrink
		if cap(s.Window) > s.WindowSize*2 {
			shrunkWindow := make([]PricePoint, len(s.Window), s.WindowSize)
			copy(shrunkWindow, s.Window)
			s.Window = shrunkWindow
		}
	}

	return oldCap, newCap
}

func (s *StockTracker) CurrentPrice() float64 {
	if len(s.Window) == 0 {
		return 0
	}

	return s.Window[len(s.Window)-1].Price
}

func (s *StockTracker) SMA() float64 {
	if len(s.Window) == 0 {
		return 0
	}

	return s.WindowSum / float64(len(s.Window))
}

func (s *StockTracker) MinMax() (minPrice float64, maxPrice float64) {
	if len(s.Window) == 0 {
		return 0, 0
	}

	minPrice, maxPrice = s.Window[0].Price, s.Window[0].Price

	for _, pricePoint := range s.Window[1:] {
		price := pricePoint.Price
		if price > maxPrice {
			maxPrice = price
		}
		if price < minPrice {
			minPrice = price
		}
	}

	return minPrice, maxPrice
}

func (s *StockTracker) String() string {
	if len(s.Window) == 0 {
		return fmt.Sprintf("[%s] No data", s.Symbol)
	}

	minPrice, maxPrice := s.MinMax()

	return fmt.Sprintf(
		"[%s] Price: %.2f | SMA: %.2f | Min: %.2f | Max: %.2f",
		s.Symbol,
		s.CurrentPrice(),
		s.SMA(),
		minPrice,
		maxPrice,
	)
}

func (s *StockTracker) SeedInitialPoint(now time.Time, rng *rand.Rand) (int, int) {
	point := PricePoint{
		Price:     s.LastGeneratedPrice,
		Volume:    nextVolume(rng),
		Timestamp: now,
	}

	return s.AddPrice(point)
}

func (s *StockTracker) NextPricePoint(now time.Time, rng *rand.Rand) PricePoint {
	changePercent := (rng.Float64() - 0.5) * 0.02
	nextPrice := s.LastGeneratedPrice * (1 + changePercent)
	if nextPrice < 1 {
		nextPrice = 1
	}

	s.LastGeneratedPrice = nextPrice

	return PricePoint{
		Price:     nextPrice,
		Volume:    nextVolume(rng),
		Timestamp: now,
	}
}

func nextVolume(rng *rand.Rand) int {
	return 100 + rng.Intn(9901)
}

func initialPriceForSymbol(symbol string) float64 {
	basePrices := map[string]float64{
		"RELIANCE": 2450.00,
		"TCS":      3900.00,
		"INFY":     1525.00,
		"HDFCBANK": 1680.00,
		"ITC":      440.00,
	}

	if price, ok := basePrices[symbol]; ok {
		return price
	}

	return 1000.00
}
