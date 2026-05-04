package garage

import (
	"regexp"
	"strings"
)

type PlateValidator func(string) bool


// Indian Number Plates
var reIndia = regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z]{1,2}[0-9]{4}$`)

func IndiaValidator(plate string) bool {
	normalized := strings.ReplaceAll(strings.ReplaceAll(plate, "-", ""), " ", "")
	return reIndia.MatchString(strings.ToUpper(normalized))
}
