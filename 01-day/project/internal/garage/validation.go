package garage

import (
	"fmt"
	"regexp"
	"strings"
)

// PlateValidator checks whether a normalized plate number is valid.
type PlateValidator func(string) bool

var reIndia = regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z]{1,2}[0-9]{4}$`)

// IndiaValidator validates normalized Indian vehicle registration numbers.
func IndiaValidator(plate string) bool {
	return reIndia.MatchString(plate)
}

// NormalizePlate converts a plate into the canonical format used by the garage.
func NormalizePlate(plate string) string {
	plate = strings.TrimSpace(plate)
	plate = strings.ToUpper(plate)
	plate = strings.ReplaceAll(plate, "-", "")
	plate = strings.ReplaceAll(plate, " ", "")
	return plate
}

// ValidateFloor validates that the requested floor exists in the garage.
func ValidateFloor(floor int, totalFloors int) error {
	if floor < 1 || floor > totalFloors {
		return fmt.Errorf("invalid floor: %d. Allowed floors are 1-%d", floor, totalFloors)
	}

	return nil
}

// ValidatePlate validates a normalized plate number against the registered rules.
func ValidatePlate(plate string, validators []PlateValidator) error {
	if plate == "" {
		return fmt.Errorf("plate number is required")
	}

	if len(validators) == 0 {
		return nil
	}

	for _, validator := range validators {
		if validator(plate) {
			return nil
		}
	}

	return fmt.Errorf("invalid plate number: %s. Expected format like GJ06MG9560", plate)
}
