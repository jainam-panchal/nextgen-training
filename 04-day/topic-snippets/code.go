package main

import (
	"fmt"
)

type Stringer interface {
	String() string
}

type Temperature struct {
	Celsius   float64
	DummyData string
}

func (t Temperature) String() string {
	return fmt.Sprintf("%.1f°C", t.Celsius)
}

func main() {
	var s Stringer = Temperature{Celsius: 36.6}

	t, ok := s.(Temperature)

	if ok {
		fmt.Println(t.DummyData)
	}
}
