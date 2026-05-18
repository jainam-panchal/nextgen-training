package pipeline

import (
	"sync"
	"time"
)

type SensorStats struct {
	Count   int64
	Sum     float64
	Average float64
}

type Stats struct {
	mu sync.Mutex

	startTime time.Time

	processedCount int64
	totalLatency   time.Duration

	alertsBySensor map[string]int64
	sensorStats    map[string]SensorStats
}

func NewStats() *Stats {
	return &Stats{
		startTime:      time.Now(),
		alertsBySensor: make(map[string]int64),
		sensorStats:    make(map[string]SensorStats),
	}
}

func (s *Stats) RecordProcessed(sensorID string, value float64, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processedCount++
	s.totalLatency += latency

	stats := s.sensorStats[sensorID]
	stats.Count++
	stats.Sum += value
	stats.Average = stats.Sum / float64(stats.Count)

	s.sensorStats[sensorID] = stats
}

func (s *Stats) RecordAlert(sensorID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.alertsBySensor[sensorID]++
}

func (s *Stats) ProcessedPerSecond() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	elapsed := time.Since(s.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}

	return float64(s.processedCount) / elapsed
}

func (s *Stats) AverageLatency() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.processedCount == 0 {
		return 0
	}

	return s.totalLatency / time.Duration(s.processedCount)
}

func (s *Stats) AlertsBySensor() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]int64, len(s.alertsBySensor))
	for sensorID, count := range s.alertsBySensor {
		result[sensorID] = count
	}

	return result
}

func (s *Stats) SensorStats() map[string]SensorStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]SensorStats, len(s.sensorStats))
	for sensorID, stats := range s.sensorStats {
		result[sensorID] = stats
	}

	return result
}

func QueueUtilization[T any](ch <-chan T) float64 {
	if cap(ch) == 0 {
		return 0
	}

	return float64(len(ch)) / float64(cap(ch)) * 100
}
