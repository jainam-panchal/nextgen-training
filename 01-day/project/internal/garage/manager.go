package garage

import (
	"fmt"
	"strings"
	"time"
)

type GarageManager struct {
	Cfg        Config
	spots      []ParkingSpot
	registry   map[string]int
	Validators []PlateValidator
}

func NewManager(cfg Config, rules []PlateValidator) *GarageManager {
	gm := &GarageManager{
		Cfg:        cfg,
		spots:      make([]ParkingSpot, cfg.TotalSpots),
		registry:   make(map[string]int),
		Validators: rules,
	}

	for i := 0; i < cfg.TotalSpots; i++ {
		gm.spots[i] = ParkingSpot{
			ID:       i + 1,
			Floor:    (i / cfg.SpotsPerFloor) + 1,
			Position: (i % cfg.SpotsPerFloor) + 1,
			Occupied: false,
		}
	}

	return gm
}

func (gm *GarageManager) Park(plate string, floor int) (*ParkingSpot, error) {
	if _, exists := gm.registry[plate]; exists {
		return nil, fmt.Errorf("vehicle %s is already parked", plate)
	}

	start := (floor - 1) * gm.Cfg.SpotsPerFloor
	end := start + gm.Cfg.SpotsPerFloor

	for i := start; i < end; i++ {
		if !gm.spots[i].Occupied {

			// Allocate
			gm.spots[i].Occupied = true
			gm.spots[i].Plate = plate
			gm.spots[i].EntryTime = time.Now()

			// Mark entry
			gm.registry[plate] = i

			return &gm.spots[i], nil
		}
	}

	return nil, fmt.Errorf("floor %d is full", floor)
}

func (gm *GarageManager) Exit(plate string) (time.Duration, int, error) {
	index, exists := gm.registry[plate]
	if !exists {
		return 0, 0, fmt.Errorf("vehicle %s is not parked", plate)
	}

	spot := &gm.spots[index]
	duration := time.Since(spot.EntryTime)
	fee := calculateStartedHourFee(duration, gm.Cfg.HourlyRate)

	spot.EntryTime = time.Time{}
	spot.Occupied = false
	spot.Plate = ""

	delete(gm.registry, plate)

	return duration, fee, nil
}

func (gm *GarageManager) Status() string {
	available := make([]int, gm.Cfg.TotalFloors)

	for _, spot := range gm.spots {
		if !spot.Occupied {
			available[spot.Floor-1]++
		}
	}

	var parts []string
	for floor := 1; floor <= gm.Cfg.TotalFloors; floor++ {
		part := fmt.Sprintf(
			"Floor %d: %d/%d available",
			floor,
			available[floor-1],
			gm.Cfg.SpotsPerFloor,
		)
		parts = append(parts, part)
	}

	return strings.Join(parts, " | ")
}

func (gm *GarageManager) Search(plate string) (*ParkingSpot, error) {
	index, exists := gm.registry[plate]
	if !exists {
		return nil, fmt.Errorf("vehicle %s not found", plate)
	}

	return &gm.spots[index], nil
}

func calculateStartedHourFee(duration time.Duration, hourlyRate int) int {
	hours := int(duration / time.Hour)
	if duration%time.Hour != 0 {
		hours++
	}
	if hours == 0 {
		hours = 1
	}

	return hours * hourlyRate
}
