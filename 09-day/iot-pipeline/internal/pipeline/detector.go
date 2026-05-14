package pipeline

import (
	"iot-pipeline/internal/models"
	"time"
)

func DetectAnomaly(reading models.SensorReading, now time.Time) (models.Alert, bool) {
	switch reading.Type {
	case models.TemperatureSensor:
		if reading.Value > 50 {
			return models.Alert{
				Reading:   reading,
				Message:   "temperature exceeded 50°C",
				Severity:  models.Critical,
				Timestamp: now,
			}, true
		}

	case models.HumiditySensor:
		if reading.Value > 95 {
			return models.Alert{
				Reading:   reading,
				Message:   "humidity exceeded 95%",
				Severity:  models.Critical,
				Timestamp: now,
			}, true
		}

	case models.MotionSensor:
		if isRestrictedHour(now) && reading.Value > 0 {
			return models.Alert{
				Reading:   reading,
				Message:   "motion detected during restricted hours",
				Severity:  models.Warning,
				Timestamp: now,
			}, true
		}
	}

	return models.Alert{}, false
}

func isRestrictedHour(now time.Time) bool {
	hour := now.Hour()
	return hour >= 22 || hour < 6
}
