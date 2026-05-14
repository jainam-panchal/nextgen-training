package pipeline

import (
	"context"
	"fmt"
	"iot-pipeline/internal/models"
	"sync"
	"time"
)

func RunProcessor(
	ctx context.Context,
	id int,
	ingestion <-chan models.SensorReading,
	alerts chan<- models.Alert,
	stats *Stats,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case reading, ok := <-ingestion:
			if !ok {
				return
			}
			fmt.Println("Processor", id, "is consuming", reading)
			latency := time.Since(reading.Timestamp)
			stats.RecordProcessed(reading.SensorID, reading.Value, latency)

			alert, detected := DetectAnomaly(reading, time.Now())
			if detected {
				select {
				case alerts <- alert:
					stats.RecordAlert(alert.Reading.SensorID)
				case <-ctx.Done():
					return
				}
			}

			time.Sleep(1 * time.Second)

		case <-time.After(5 * time.Second):
			fmt.Println("Processor", id, "has been idle for 5 seconds")
		}
	}
}

//func RunProcessor(
//	ctx context.Context,
//	id int,
//	ingestion <-chan models.SensorReading,
//	alerts chan<- models.Alert,
//	wg *sync.WaitGroup,
//) {
//	defer wg.Done()
//
//	for reading := range ingestion {
//		fmt.Println("Processor", id, "is consuming", reading)
//
//		alert, detected := DetectAnomaly(reading, time.Now())
//		if detected {
//			select {
//			case alerts <- alert:
//			case <-ctx.Done():
//				return
//			}
//		}
//
//		time.Sleep(1 * time.Second)
//	}
//}
