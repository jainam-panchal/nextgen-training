package pipeline

import (
	"testing"
	"time"

	"iot-pipeline/internal/models"
)

func TestDetectAnomaly(t *testing.T) {
	restrictedTime := time.Date(2026, 1, 1, 23, 0, 0, 0, time.UTC)
	normalTime := time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		reading      models.SensorReading
		now          time.Time
		wantAnomaly  bool
		wantSeverity models.AlertSeverity
	}{
		{
			name: "high temperature creates critical alert",
			reading: models.SensorReading{
				SensorID: "temp-01",
				Type:     models.TemperatureSensor,
				Value:    51,
			},
			now:          normalTime,
			wantAnomaly:  true,
			wantSeverity: models.Critical,
		},
		{
			name: "temperature exactly at threshold does not alert",
			reading: models.SensorReading{
				SensorID: "temp-01",
				Type:     models.TemperatureSensor,
				Value:    50,
			},
			now:         normalTime,
			wantAnomaly: false,
		},
		{
			name: "normal temperature does not alert",
			reading: models.SensorReading{
				SensorID: "temp-01",
				Type:     models.TemperatureSensor,
				Value:    25,
			},
			now:         normalTime,
			wantAnomaly: false,
		},
		{
			name: "high humidity creates critical alert",
			reading: models.SensorReading{
				SensorID: "humidity-01",
				Type:     models.HumiditySensor,
				Value:    96,
			},
			now:          normalTime,
			wantAnomaly:  true,
			wantSeverity: models.Critical,
		},
		{
			name: "humidity exactly at threshold does not alert",
			reading: models.SensorReading{
				SensorID: "humidity-01",
				Type:     models.HumiditySensor,
				Value:    95,
			},
			now:         normalTime,
			wantAnomaly: false,
		},
		{
			name: "normal humidity does not alert",
			reading: models.SensorReading{
				SensorID: "humidity-01",
				Type:     models.HumiditySensor,
				Value:    60,
			},
			now:         normalTime,
			wantAnomaly: false,
		},
		{
			name: "motion during restricted hours creates warning alert",
			reading: models.SensorReading{
				SensorID: "motion-01",
				Type:     models.MotionSensor,
				Value:    1,
			},
			now:          restrictedTime,
			wantAnomaly:  true,
			wantSeverity: models.Warning,
		},
		{
			name: "motion during normal hours does not alert",
			reading: models.SensorReading{
				SensorID: "motion-01",
				Type:     models.MotionSensor,
				Value:    1,
			},
			now:         normalTime,
			wantAnomaly: false,
		},
		{
			name: "no motion during restricted hours does not alert",
			reading: models.SensorReading{
				SensorID: "motion-01",
				Type:     models.MotionSensor,
				Value:    0,
			},
			now:         restrictedTime,
			wantAnomaly: false,
		},
		{
			name: "pressure reading does not alert",
			reading: models.SensorReading{
				SensorID: "pressure-01",
				Type:     models.PressureSensor,
				Value:    1013,
			},
			now:         normalTime,
			wantAnomaly: false,
		},
		{
			name: "light reading does not alert",
			reading: models.SensorReading{
				SensorID: "light-01",
				Type:     models.LightSensor,
				Value:    500,
			},
			now:         normalTime,
			wantAnomaly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert, gotAnomaly := DetectAnomaly(tt.reading, tt.now)

			if gotAnomaly != tt.wantAnomaly {
				t.Fatalf("DetectAnomaly() anomaly = %v, want %v", gotAnomaly, tt.wantAnomaly)
			}

			if !tt.wantAnomaly {
				return
			}

			if alert.Reading.SensorID != tt.reading.SensorID {
				t.Fatalf("alert sensor ID = %q, want %q", alert.Reading.SensorID, tt.reading.SensorID)
			}

			if alert.Severity != tt.wantSeverity {
				t.Fatalf("alert severity = %q, want %q", alert.Severity, tt.wantSeverity)
			}

			if alert.Message == "" {
				t.Fatal("alert message should not be empty")
			}

			if !alert.Timestamp.Equal(tt.now) {
				t.Fatalf("alert timestamp = %v, want %v", alert.Timestamp, tt.now)
			}
		})
	}
}

func TestIsRestrictedHour(t *testing.T) {
	tests := []struct {
		name string
		hour int
		want bool
	}{
		{name: "midnight is restricted", hour: 0, want: true},
		{name: "5am is restricted", hour: 5, want: true},
		{name: "6am is not restricted", hour: 6, want: false},
		{name: "2pm is not restricted", hour: 14, want: false},
		{name: "9pm is not restricted", hour: 21, want: false},
		{name: "10pm is restricted", hour: 22, want: true},
		{name: "11pm is restricted", hour: 23, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, 1, 1, tt.hour, 0, 0, 0, time.UTC)

			got := isRestrictedHour(now)

			if got != tt.want {
				t.Fatalf("isRestrictedHour(%d) = %v, want %v", tt.hour, got, tt.want)
			}
		})
	}
}
