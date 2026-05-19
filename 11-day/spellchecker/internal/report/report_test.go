package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"spellchecker/internal/ranking"
)

func TestMarshalJSON(t *testing.T) {
	r := SpellCheckReport{
		TotalWords:      10,
		MisspelledWords: []string{"wrod"},
		Corrections: []Correction{
			{
				Word: "wrod",
				Suggestions: []ranking.Suggestion{
					{Word: "word", Distance: 1, Frequency: 2},
				},
			},
		},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got := int(decoded["total_words"].(float64)); got != 10 {
		t.Fatalf("total_words mismatch: got %d", got)
	}
	if got := int(decoded["misspelled"].(float64)); got != 1 {
		t.Fatalf("misspelled mismatch: got %d", got)
	}
}

func TestWriteToFile(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "report.json")
	r := SpellCheckReport{
		TotalWords:      1,
		MisspelledWords: []string{"x"},
		Corrections:     []Correction{{Word: "x"}},
	}

	if err := WriteToFile(outPath, r); err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected report file to be non-empty")
	}
}
