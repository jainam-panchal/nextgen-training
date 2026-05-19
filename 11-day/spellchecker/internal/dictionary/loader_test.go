package dictionary

import (
	"path/filepath"
	"testing"

	"spellchecker/trie"
)

func testDictionaryPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "data", "words.txt")
}

func TestLoadFromFile(t *testing.T) {
	dictPath := testDictionaryPath(t)

	dictTrie := trie.NewTrie()
	inserted, err := LoadFromFile(dictPath, dictTrie)
	if err != nil {
		t.Fatalf("LoadFromFile returned unexpected error: %v", err)
	}

	if inserted == 0 {
		t.Fatal("expected inserted count > 0")
	}

	if !dictTrie.Search("program") {
		t.Fatal("expected 'program' to be loaded")
	}

	if got := dictTrie.Frequency("program"); got <= 0 {
		t.Fatalf("expected frequency > 0 for 'program', got %d", got)
	}
}

func TestLoadFromFileNilTrie(t *testing.T) {
	dictPath := testDictionaryPath(t)

	_, err := LoadFromFile(dictPath, nil)
	if err == nil {
		t.Fatal("expected error for nil trie")
	}
}

func TestLoadWordsAndTrie(t *testing.T) {
	dictPath := testDictionaryPath(t)

	dictTrie := trie.NewTrie()
	words, inserted, err := LoadWordsAndTrie(dictPath, dictTrie)
	if err != nil {
		t.Fatalf("LoadWordsAndTrie returned unexpected error: %v", err)
	}

	if inserted == 0 {
		t.Fatal("expected inserted count > 0")
	}
	if len(words) != inserted {
		t.Fatalf("expected words len == inserted (%d), got %d", inserted, len(words))
	}
	if words[0] == "" {
		t.Fatal("expected first loaded word to be non-empty")
	}
}
