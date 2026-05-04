package cli

import (
	"fmt"
	"strconv"

	"jainam-panchal/nextgen-tranining/gms/internal/garage"
)

type commandHandler func(m *garage.GarageManager, args ...string)

func handlePark(m *garage.GarageManager, args ...string) {
	if len(args) != 2 {
		fmt.Println("Usage: park <plate> <floor>")
		return
	}

	plate := garage.NormalizePlate(args[0])
	if err := garage.ValidatePlate(plate, m.Validators); err != nil {
		fmt.Println(err.Error())
		return
	}

	floor, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Invalid floor. Please enter a number.")
		return
	}

	if err := garage.ValidateFloor(floor, m.Cfg.TotalFloors); err != nil {
		fmt.Println(err.Error())
		return
	}

	spot, err := m.Park(plate, floor)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Vehicle %s parked at Floor %d, Spot %d\n", plate, spot.Floor, spot.ID)
}

func handleExit(m *garage.GarageManager, args ...string) {
	if len(args) != 1 {
		fmt.Println("Usage: exit <plate>")
		return
	}

	plate := garage.NormalizePlate(args[0])
	if err := garage.ValidatePlate(plate, m.Validators); err != nil {
		fmt.Println(err.Error())
		return
	}

	duration, fees, err := m.Exit(plate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Vehicle %s exited. Duration: %s. Fee: Rs %d\n", plate, formatDuration(duration), fees)
}

func handleStatus(m *garage.GarageManager, args ...string) {
	if len(args) != 0 {
		fmt.Println("Usage: status")
		return
	}

	fmt.Println(m.Status())
}

func handleSearch(m *garage.GarageManager, args ...string) {
	if len(args) != 1 {
		fmt.Println("Usage: find <plate>")
		return
	}

	plate := garage.NormalizePlate(args[0])
	if err := garage.ValidatePlate(plate, m.Validators); err != nil {
		fmt.Println(err.Error())
		return
	}

	spot, err := m.Search(plate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Vehicle %s is parked at Floor %d, Spot %d\n", plate, spot.Floor, spot.ID)
}
