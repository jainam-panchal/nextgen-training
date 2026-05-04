package cli

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{name: "seconds only", duration: 15 * time.Second, want: "15s"},
		{name: "minutes and seconds", duration: time.Minute + 5*time.Second, want: "1m 5s"},
		{name: "hours and minutes", duration: 2*time.Hour + 15*time.Minute, want: "2h 15m"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Fatalf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}
