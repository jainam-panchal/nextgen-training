package garage

import (
	"testing"
	"time"
)

func TestCalculateStartedHourFee(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		rate     int
		want     int
	}{
		{name: "partial hour rounds up", duration: 15 * time.Minute, rate: 20, want: 20},
		{name: "exact hour stays exact", duration: time.Hour, rate: 20, want: 20},
		{name: "multiple hours rounds up", duration: 2*time.Hour + 15*time.Minute, rate: 20, want: 60},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := calculateStartedHourFee(tt.duration, tt.rate)
			if got != tt.want {
				t.Fatalf("calculateStartedHourFee(%v, %d) = %d, want %d", tt.duration, tt.rate, got, tt.want)
			}
		})
	}
}
