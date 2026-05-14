### IoT Sensor Pipeline

Components:
Sensors (10 goroutines): produce SensorReading every 100-500ms
Ingestion Channel: buffered chan of SensorReading (capacity: 100)
Processors (3 goroutines): read from ingestion, compute stats, detect anomalies
Alert Channel: buffered chan of Alert (capacity: 50)
Alert Handler (1 goroutine): deduplicates alerts, logs them

Orchestration:
context.Context: cancellation signal flows from main → all goroutines
sync.WaitGroup: main waits for all goroutines to exit cleanly
Stats goroutine: periodic throughput + queue utilization reporting
