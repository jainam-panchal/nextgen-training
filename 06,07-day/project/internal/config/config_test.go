package config

import "testing"

// Tests default config values used by dispatcher.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CellSizeDegrees != 0.01 {
		t.Fatalf("expected CellSizeDegrees=0.01, got %f", cfg.CellSizeDegrees)
	}
	if cfg.MaxPickupDistanceKm != 5.0 {
		t.Fatalf("expected MaxPickupDistanceKm=5.0, got %f", cfg.MaxPickupDistanceKm)
	}
	if cfg.BaseFare != 50.0 {
		t.Fatalf("expected BaseFare=50.0, got %f", cfg.BaseFare)
	}
	if cfg.FarePerKm != 12.0 {
		t.Fatalf("expected FarePerKm=12.0, got %f", cfg.FarePerKm)
	}
}

// Tests NeighborRing basic sanity with default values.
func TestNeighborRing(t *testing.T) {
	cfg := DefaultConfig()
	ring := cfg.NeighborRing()

	if ring <= 0 {
		t.Fatalf("expected positive neighbor ring, got %d", ring)
	}
	// 0.01 degree ~1.11km, 5km radius -> ceil(4.5) == 5
	if ring != 5 {
		t.Fatalf("expected neighbor ring 5, got %d", ring)
	}
}
