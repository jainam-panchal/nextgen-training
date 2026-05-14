package pipeline

import (
	"context"
	"fmt"
	"iot-pipeline/internal/models"
	"math/rand"
	"sync"
	"time"
)

func RunSensor(
	ctx context.Context,
	config SensorConfig,
	out chan<- models.SensorReading,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		randInterval := generateInterval(config.MinInterval, config.MaxInterval)
		timer := time.NewTimer(randInterval)

		select {
		case <-ctx.Done():
			timer.Stop()
			return

		case <-timer.C:
			reading := models.SensorReading{
				SensorID:  config.ID,
				Type:      config.Type,
				Value:     generateNumber(config.MinValue, config.MaxValue),
				Unit:      config.Unit,
				Location:  config.Location,
				Timestamp: time.Now(),
			}

			utilization := float64(len(out)) / float64(cap(out)) * 100
			if utilization >= 85 {
				fmt.Printf("Queue %.0f%% full — applying backpressure\n", utilization)
			}

			select {
			case <-ctx.Done():
				return
			case out <- reading:
			}
		}
	}
}

func generateNumber(minValue float64, maxValue float64) float64 {
	if maxValue <= minValue {
		return minValue
	}

	return rand.Float64()*(maxValue-minValue) + minValue
}

func generateInterval(minInterval time.Duration, maxInterval time.Duration) time.Duration {
	if maxInterval <= minInterval {
		return minInterval
	}

	delta := int64(maxInterval - minInterval)
	return minInterval + time.Duration(rand.Int63n(delta))
}
