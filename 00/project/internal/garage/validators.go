package garage

import (
	"fmt"
	"regexp"
	"strings"
)

type PlateValidator func(string) bool

// Indian Number Plates
var reIndia = regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z]{1,2}[0-9]{4}$`)

func IndiaValidator(plate string) bool {
	return reIndia.MatchString(plate)
}

func NormalizePlate(plate string) string {
	plate = strings.TrimSpace(plate)
	plate = strings.ToUpper(plate)
	plate = strings.ReplaceAll(plate, "-", "")
	plate = strings.ReplaceAll(plate, " ", "")
	return plate
}

func ValidateFloor(floor int, totalFloors int) error {
	if floor < 1 || floor > totalFloors {
		return fmt.Errorf("invalid floor: %d. Allowed floors are 1-%d", floor, totalFloors)
	}

	return nil
}

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

	return fmt.Errorf("invalid plate number: %s", plate)
}
