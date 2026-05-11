package models

type Location struct {
	Lat float64
	Lng float64
}

type BlockID struct {
	LatBucket int
	LngBucket int
}
