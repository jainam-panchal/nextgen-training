package garage

type GarageManager struct {
	cfg        Config
	spots      []ParkingSpot
	registar   map[string]string
	validators []PlateValidator
}

func NewManager(cfg Config, rules []PlateValidator) *GarageManager {
	gm := &GarageManager{
		cfg:        cfg,
		spots:      make([]ParkingSpot, cfg.TotalSpots),
		registar:   make(map[string]string),
		validators: rules,
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

func (gm *GarageManager) isPlateValid(plate string) bool {
	if len(gm.validators) == 0 {
		return true
	}
	for _, v := range gm.validators {
		if v(plate) {
			return true
		}
	}
	return false
}
