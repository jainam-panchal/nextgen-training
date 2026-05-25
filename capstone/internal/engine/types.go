package engine

import "time"

// VehicleType distinguishes normal traffic from emergency vehicles.
type VehicleType string

const (
	VehicleTypeNormal    VehicleType = "normal"    // standard vehicle, uses normal routing
	VehicleTypeEmergency VehicleType = "emergency" // emergency vehicle, gets corridor + preemption
)

// Phase represents the state of a traffic signal light at one intersection.
type Phase string

const (
	PhaseGreenNS  Phase = "Green-NS"  // green light for north-south direction
	PhaseYellowNS Phase = "Yellow-NS" // yellow (transitioning NS→EW)
	PhaseGreenEW  Phase = "Green-EW"  // green light for east-west direction
	PhaseYellowEW Phase = "Yellow-EW" // yellow (transitioning EW→NS)
)

// VehicleState tracks where a vehicle is in its lifecycle.
type VehicleState string

const (
	VehicleStateAtIntersection VehicleState = "at_intersection" // stopped at (or just arrived at) an intersection
	VehicleStateWaitingSignal  VehicleState = "waiting_signal"  // waiting for the signal to turn green for our direction
	VehicleStateOnRoad         VehicleState = "on_road"         // traversing a road segment between intersections
	VehicleStateArrived        VehicleState = "arrived"         // reached the final destination intersection
)

// Config holds all tuning parameters for the simulation engine.
type Config struct {
	// Number of intersections (nodes) in the graph.
	NumIntersections int

	// Signal timing expressed in ticks.
	MinGreenTicks int // minimum green duration (used when no vehicles are queued)
	MaxGreenTicks int // maximum green duration (used when queues are heavily imbalanced)
	YellowTicks   int // duration of the yellow (transition) phase

	// Congestion history window in ticks. Determines how many samples
	// the rolling average congestion considers.
	HistoryWindow int

	// Emergency preemption: how many ticks the corridor remains active
	// and signals stay preempted.
	PreemptionTicks int

	// Simulation tick duration (wall-clock time per tick) for RealClock.
	TickDuration time.Duration
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() Config {
	return Config{
		NumIntersections: 20,
		MinGreenTicks:    10,
		MaxGreenTicks:    30,
		YellowTicks:      3,
		HistoryWindow:    600,
		PreemptionTicks:  60,
		TickDuration:     1 * time.Second,
	}
}
