package dictionary

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"spellchecker/trie"
)

const maxScannerBuffer = 1024 * 1024

func LoadFromFile(path string, dictTrie *trie.Trie) (int, error) {
	if dictTrie == nil {
		return 0, fmt.Errorf("dictionary trie cannot be nil")
	}

	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open dictionary file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, maxScannerBuffer)

	inserted := 0
	for scanner.Scan() {
		word := normalizeWord(scanner.Text())
		if word == "" {
			continue
		}
		if strings.HasPrefix(word, "#") {
			continue
		}

		dictTrie.Insert(word, 1)
		inserted++
	}

	if err := scanner.Err(); err != nil {
		return inserted, fmt.Errorf("scan dictionary file %q: %w", path, err)
	}

	return inserted, nil
}

func LoadWordsAndTrie(path string, dictTrie *trie.Trie) ([]string, int, error) {
	if dictTrie == nil {
		return nil, 0, fmt.Errorf("dictionary trie cannot be nil")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("open dictionary file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, maxScannerBuffer)

	inserted := 0
	words := make([]string, 0, 1024)
	for scanner.Scan() {
		word := normalizeWord(scanner.Text())
		if word == "" {
			continue
		}
		if strings.HasPrefix(word, "#") {
			continue
		}

		dictTrie.Insert(word, 1)
		words = append(words, word)
		inserted++
	}

	if err := scanner.Err(); err != nil {
		return words, inserted, fmt.Errorf("scan dictionary file %q: %w", path, err)
	}

	return words, inserted, nil
}

func normalizeWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}
