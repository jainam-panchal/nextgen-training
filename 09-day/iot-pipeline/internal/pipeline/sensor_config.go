package pipeline

import (
	"time"

	"iot-pipeline/internal/models"
)

type SensorConfig struct {
	ID          string
	Type        models.SensorType
	Unit        string
	Location    string
	MinValue    float64
	MaxValue    float64
	MinInterval time.Duration
	MaxInterval time.Duration
}

var AllSensorConfigs = []SensorConfig{
	{ID: "temp-01", Type: models.TemperatureSensor, Unit: "C", Location: "room-a", MinValue: 18, MaxValue: 60, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},
	{ID: "temp-02", Type: models.TemperatureSensor, Unit: "C", Location: "room-b", MinValue: 18, MaxValue: 60, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},

	{ID: "humidity-01", Type: models.HumiditySensor, Unit: "%", Location: "room-a", MinValue: 30, MaxValue: 100, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},
	{ID: "humidity-02", Type: models.HumiditySensor, Unit: "%", Location: "room-b", MinValue: 30, MaxValue: 100, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},

	{ID: "motion-01", Type: models.MotionSensor, Unit: "state", Location: "entrance", MinValue: 0, MaxValue: 1, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},
	{ID: "motion-02", Type: models.MotionSensor, Unit: "state", Location: "corridor", MinValue: 0, MaxValue: 1, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},

	{ID: "pressure-01", Type: models.PressureSensor, Unit: "hPa", Location: "outdoor", MinValue: 980, MaxValue: 1040, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},
	{ID: "pressure-02", Type: models.PressureSensor, Unit: "hPa", Location: "server-room", MinValue: 980, MaxValue: 1040, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},

	{ID: "light-01", Type: models.LightSensor, Unit: "lux", Location: "office", MinValue: 0, MaxValue: 1000, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},
	{ID: "light-02", Type: models.LightSensor, Unit: "lux", Location: "warehouse", MinValue: 0, MaxValue: 1000, MinInterval: 100 * time.Millisecond, MaxInterval: 500 * time.Millisecond},
}
