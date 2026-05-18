package pipeline

import (
	"fmt"
	"sync"
	"time"

	"iot-pipeline/internal/models"
)

func RunAlertHandler(
	alerts <-chan models.Alert,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	lastAlertAt := make(map[string]time.Time)
	cooldown := 60 * time.Second

	for {
		select {
		case alert, ok := <-alerts:
			if !ok {
				return
			}

			sensorID := alert.Reading.SensorID
			now := time.Now()

			lastSeen, exists := lastAlertAt[sensorID]
			if exists && now.Sub(lastSeen) < cooldown {
				fmt.Println("duplicate alert suppressed:", sensorID)
				continue
			}

			lastAlertAt[sensorID] = now
			fmt.Println("alert has been detected:", alert)

		}
	}
}
