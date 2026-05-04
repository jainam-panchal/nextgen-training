package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"jainam-panchal/nextgen-tranining/gms/internal/garage"
)

func getEnvInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)

	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}

	return defaultValue
}

func loadConfig() garage.Config {
	return garage.Config{
		TotalFloors:   getEnvInt("PARKING_TOTAL_FLOORS", 5),
		SpotsPerFloor: getEnvInt("PARKING_SPOTS_PER_FLOOR", 100),
		HourlyRate:    getEnvInt("PARKING_HOURLY_RATE", 20),
		TotalSpots:    getEnvInt("PARKING_TOTAL_FLOORS", 5) * getEnvInt("PARKING_SPOTS_PER_FLOOR", 100),
	}
}

func main() {
	cfg := loadConfig()

	rules := []garage.PlateValidator{
		garage.IndiaValidator,
	}

	manager := garage.NewManager(cfg, rules)
	scanner := bufio.NewScanner(os.Stdin)

	handlers := map[string]commandHandler{
		"P": handlePark,   // Validates and parks a vehicle on the requested floor.
		"E": handleExit,   // Process exit, clears the spot, and calculates the fee.
		"S": handleStatus, // Displays current available spots across all floors.
		"F": handleSearch, // Finds the specific floor and spot location for a plate.
	}

	fmt.Println("=== Parking Garage Management System ===")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  P <plate> <floor>   Park a vehicle on the requested floor")
	fmt.Println("  E <plate>           Exit a parked vehicle and calculate the fee")
	fmt.Println("  F <plate>           Find a parked vehicle by plate number")
	fmt.Println("  S                   Show floor-wise availability summary")
	fmt.Println("  Q                   Quit the application")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  P MH12AB1234 2")
	fmt.Println("  E MH12AB1234")
	fmt.Println("  F MH12AB1234")
	fmt.Println("  S")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Plates may be entered with or without spaces/hyphens.")
	fmt.Println("  - Parking is assigned to the first available spot on the requested floor.")
	fmt.Println("  - Exit fee is charged at Rs 20 per started hour.")
	fmt.Println()

	for {
		fmt.Print("garage> ")

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToUpper(parts[0])

		args := parts[1:] // Take all other params

		if cmd == "Q" {
			fmt.Println("Taking the system offline.")
			break
		}

		if handler, exits := handlers[cmd]; exits {
			handler(manager, args...) // the variadic call
		} else {
			fmt.Printf("Unknown command: %s. Available: P, E, S, F, Q\n", cmd)
		}
	}
}
