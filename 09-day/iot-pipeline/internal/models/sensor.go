package models

import "time"

type SensorReading struct {
	SensorID  string
	Type      SensorType
	Value     float64
	Unit      string
	Location  string
	Timestamp time.Time
}

type Alert struct {
	Reading   SensorReading
	Message   string
	Severity  AlertSeverity
	Timestamp time.Time
}

type SensorType string
type AlertSeverity string

const (
	HumiditySensor    SensorType = "humidity"
	TemperatureSensor SensorType = "temperature"
	MotionSensor      SensorType = "motion"
	LightSensor       SensorType = "light"
	PressureSensor    SensorType = "pressure"

	Info     AlertSeverity = "info"
	Warning  AlertSeverity = "warning"
	Critical AlertSeverity = "critical"
)
