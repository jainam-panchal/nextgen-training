package main

import (
	"fmt"

	"jainam-panchal/nextgen-tranining/gms/internal/garage"
)

type commandHandler func(m *garage.GarageManager, args ...string)

func handlePark(m *garage.GarageManager, args ...string) {
	fmt.Println("Handle Parking!")
}

func handleExit(m *garage.GarageManager, args ...string) {
	fmt.Println("Handle Exit!")
}

func handleStatus(m *garage.GarageManager, args ...string) {
	fmt.Println("Handle Status!")
}

func handleSearch(m *garage.GarageManager, args ...string) {
	fmt.Println("Handle Search!")
}
