package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"jainam-panchal/nextgen-tranining/gms/internal/garage"
)

// Run starts the garage CLI loop.
func Run() error {
	cfg := loadConfig()

	rules := []garage.PlateValidator{
		garage.IndiaValidator,
	}

	manager := garage.NewManager(cfg, rules)
	scanner := bufio.NewScanner(os.Stdin)

	handlers := map[string]commandHandler{
		"P": handlePark,
		"E": handleExit,
		"S": handleStatus,
		"F": handleSearch,
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
			return scanner.Err()
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToUpper(parts[0])
		args := parts[1:]

		if cmd == "Q" {
			fmt.Println("Taking the system offline.")
			return nil
		}

		if handler, exists := handlers[cmd]; exists {
			handler(manager, args...)
			continue
		}

		fmt.Printf("Unknown command: %s. Available: P, E, S, F, Q\n", cmd)
	}
}

func getEnvInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)

	val, err := strconv.Atoi(valStr)
	if err == nil {
		return val
	}

	return defaultValue
}

func loadConfig() garage.Config {
	totalFloors := getEnvInt("PARKING_TOTAL_FLOORS", 5)
	spotsPerFloor := getEnvInt("PARKING_SPOTS_PER_FLOOR", 100)

	return garage.Config{
		TotalFloors:   totalFloors,
		SpotsPerFloor: spotsPerFloor,
		HourlyRate:    getEnvInt("PARKING_HOURLY_RATE", 20),
		TotalSpots:    totalFloors * spotsPerFloor,
	}
}
