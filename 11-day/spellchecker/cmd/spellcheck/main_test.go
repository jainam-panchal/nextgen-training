package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunEndToEnd(t *testing.T) {
	dictPath := filepath.Join("..", "..", "data", "unix-words.txt")
	inputPath := filepath.Join(t.TempDir(), "input.txt")
	outputPath := filepath.Join(t.TempDir(), "report.json")

	input := "this is a progrm with typo"
	if err := os.WriteFile(inputPath, []byte(input), 0o600); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	if err := run(dictPath, inputPath, outputPath, 5); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	out, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read report failed: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected report output")
	}
}
