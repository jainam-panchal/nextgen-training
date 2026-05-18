package main

import (
	"context"
	"fmt"
	"iot-pipeline/internal/models"
	"iot-pipeline/internal/pipeline"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats := pipeline.NewStats()
	readings := make(chan models.SensorReading, 10)
	alerts := make(chan models.Alert, 50)

	sensorWg := sync.WaitGroup{}
	processorWg := sync.WaitGroup{}
	alertWg := sync.WaitGroup{}

	alertWg.Add(1)
	go pipeline.RunAlertHandler(alerts, &alertWg)

	for i := 0; i < 3; i++ {
		processorWg.Add(1)
		go pipeline.RunProcessor(ctx, i+1, readings, alerts, stats, &processorWg)
	}

	for _, config := range pipeline.AllSensorConfigs {
		config := config
		sensorWg.Add(1)
		go pipeline.RunSensor(ctx, config, readings, &sensorWg)
	}

	<-ctx.Done()
	fmt.Println("ctx done")

	sensorWg.Wait()
	fmt.Println("sensors done")

	close(readings)
	fmt.Println("readings closed")

	processorWg.Wait()
	fmt.Println("processors done")

	close(alerts)
	fmt.Println("alerts closed")

	alertWg.Wait()
	fmt.Println("alert handler done")

	fmt.Printf("processed/sec: %.2f\n", stats.ProcessedPerSecond())
	fmt.Println("avg latency:", stats.AverageLatency())
	fmt.Println("alerts by sensor:", stats.AlertsBySensor())
	fmt.Println("sensor stats:", stats.SensorStats())
	fmt.Printf("readings queue utilization: %.2f%%\n", pipeline.QueueUtilization(readings))
	fmt.Printf("alerts queue utilization: %.2f%%\n", pipeline.QueueUtilization(alerts))
}
