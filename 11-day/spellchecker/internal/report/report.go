package report

import (
	"encoding/json"
	"fmt"
	"os"

	"spellchecker/internal/ranking"
)

type Correction struct {
	Word        string               `json:"word"`
	Suggestions []ranking.Suggestion `json:"suggestions"`
}

type SpellCheckReport struct {
	TotalWords      int
	MisspelledWords []string
	Corrections     []Correction
}

func (report SpellCheckReport) MarshalJSON() ([]byte, error) {
	type jsonReport struct {
		TotalWords  int          `json:"total_words"`
		Misspelled  int          `json:"misspelled"`
		Corrections []Correction `json:"corrections"`
	}

	payload := jsonReport{
		TotalWords:  report.TotalWords,
		Misspelled:  len(report.MisspelledWords),
		Corrections: report.Corrections,
	}

	return json.Marshal(payload)
}

func WriteToFile(path string, report SpellCheckReport) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create report file %q: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("encode report JSON: %w", err)
	}

	return nil
}
