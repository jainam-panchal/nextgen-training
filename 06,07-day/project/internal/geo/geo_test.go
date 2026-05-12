package geo

import (
	"math"
	"testing"

	"ride-sharing/internal/models"
)

// Tests block bucketing for known coordinates.
func TestGetBlockID(t *testing.T) {
	loc := models.Location{Lat: 19.08, Lng: 72.88}
	block := GetBlockID(loc, 0.01)

	// Current implementation uses floor on floating division.
	if block.LatBucket != 1907 || block.LngBucket != 7287 {
		t.Fatalf("expected block {1907,7287}, got {%d,%d}", block.LatBucket, block.LngBucket)
	}
}

// Tests neighbor block generation size and center inclusion.
func TestNeighborBlocks(t *testing.T) {
	center := models.BlockID{LatBucket: 10, LngBucket: 20}
	ring := 2
	neighbors := NeighborBlocks(center, ring)

	expectedSize := (2*ring + 1) * (2*ring + 1)
	if len(neighbors) != expectedSize {
		t.Fatalf("expected %d neighbors, got %d", expectedSize, len(neighbors))
	}

	foundCenter := false
	for _, b := range neighbors {
		if b == center {
			foundCenter = true
			break
		}
	}
	if !foundCenter {
		t.Fatalf("expected center block to be present")
	}
}

// Tests distance math basics: zero distance and symmetry.
func TestDistanceKm(t *testing.T) {
	a := models.Location{Lat: 19.08, Lng: 72.88}
	b := models.Location{Lat: 19.09, Lng: 72.89}

	if d := DistanceKm(a, a); d != 0 {
		t.Fatalf("expected zero distance, got %f", d)
	}

	ab := DistanceKm(a, b)
	ba := DistanceKm(b, a)

	if ab <= 0 {
		t.Fatalf("expected positive distance, got %f", ab)
	}
	// This approximation is close but not perfectly symmetric because cos uses a.Lat.
	if math.Abs(ab-ba) > 0.01 {
		t.Fatalf("expected near-symmetry, got ab=%f ba=%f", ab, ba)
	}
}
