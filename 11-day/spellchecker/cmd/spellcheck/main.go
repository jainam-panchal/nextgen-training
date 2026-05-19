package main

import (
	"flag"
	"fmt"
	"os"

	"spellchecker/internal/dictionary"
	"spellchecker/internal/report"
	"spellchecker/internal/spellcheck"
	"spellchecker/trie"
)

func main() {
	dictPath := flag.String("dict", "data/dictionary.txt", "path to dictionary file")
	inputPath := flag.String("in", "data/input.txt", "path to input text file")
	outputPath := flag.String("out", "data/report.json", "path to output report file")
	limit := flag.Int("limit", 5, "max suggestions per misspelled word")
	flag.Parse()

	if err := run(*dictPath, *inputPath, *outputPath, *limit); err != nil {
		fmt.Fprintf(os.Stderr, "spellcheck failed: %v\n", err)
		os.Exit(1)
	}
}

func run(dictPath string, inputPath string, outputPath string, limit int) error {
	if limit <= 0 {
		limit = 5
	}

	dictTrie := trie.NewTrie()
	dictWords, _, err := dictionary.LoadWordsAndTrie(dictPath, dictTrie)
	if err != nil {
		return err
	}

	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input file %q: %w", inputPath, err)
	}

	inputText := string(inputBytes)
	tokens := spellcheck.Tokenize(inputText)
	misspelled := spellcheck.MisspelledWords(inputText, dictTrie)

	corrections := make([]report.Correction, 0, len(misspelled))
	for _, word := range misspelled {
		suggestions := spellcheck.DidYouMean(word, dictWords, dictTrie, limit)
		corrections = append(corrections, report.Correction{
			Word:        word,
			Suggestions: suggestions,
		})
	}

	result := report.SpellCheckReport{
		TotalWords:      len(tokens),
		MisspelledWords: misspelled,
		Corrections:     corrections,
	}

	if err := report.WriteToFile(outputPath, result); err != nil {
		return err
	}

	return nil
}
