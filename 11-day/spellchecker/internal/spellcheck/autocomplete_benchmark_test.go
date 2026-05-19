package spellcheck

import (
	"path/filepath"
	"testing"

	"spellchecker/internal/dictionary"
	"spellchecker/trie"
)

func BenchmarkAutoCompleteLargeDictionary(b *testing.B) {
	dictPath := filepath.Join("..", "..", "data", "unix-words.txt")
	dictTrie := trie.NewTrie()

	if _, err := dictionary.LoadFromFile(dictPath, dictTrie); err != nil {
		b.Fatalf("load dictionary: %v", err)
	}

	prefixes := []string{"pro", "prog", "pre", "trans", "micro", "inter"}
	limit := 5
	prefixIdx := 0

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = AutoComplete(prefixes[prefixIdx], dictTrie, limit)
		prefixIdx++
		if prefixIdx == len(prefixes) {
			prefixIdx = 0
		}
	}
}
