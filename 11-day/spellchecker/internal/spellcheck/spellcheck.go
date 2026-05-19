package spellcheck

import (
	"strings"
	"unicode"

	"spellchecker/internal/distance"
	"spellchecker/internal/ranking"
	"spellchecker/trie"
)

func Tokenize(text string) []string {
	if strings.TrimSpace(text) == "" {
		return []string{}
	}

	words := make([]string, 0)
	current := make([]rune, 0, 16)

	flush := func() {
		if len(current) == 0 {
			return
		}
		words = append(words, strings.ToLower(string(current)))
		current = current[:0]
	}

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) {
			current = append(current, r)
			continue
		}
		flush()
	}
	flush()

	return words
}

func MisspelledWords(text string, dictTrie *trie.Trie) []string {
	if dictTrie == nil {
		return []string{}
	}

	tokens := Tokenize(text)
	seen := make(map[string]struct{}, len(tokens))
	misspelled := make([]string, 0)

	for _, token := range tokens {
		if token == "" {
			continue
		}
		if _, exists := seen[token]; exists {
			continue
		}
		seen[token] = struct{}{}

		if !dictTrie.Search(token) {
			misspelled = append(misspelled, token)
		}
	}

	return misspelled
}

func AutoComplete(prefix string, dictTrie *trie.Trie, limit int) []trie.Suggestion {
	if dictTrie == nil {
		return []trie.Suggestion{}
	}
	return dictTrie.AutoComplete(strings.ToLower(strings.TrimSpace(prefix)), limit)
}

func DidYouMean(word string, dictWords []string, dictTrie *trie.Trie, limit int) []ranking.Suggestion {
	if dictTrie == nil || limit <= 0 {
		return []ranking.Suggestion{}
	}

	normalized := strings.ToLower(strings.TrimSpace(word))
	if normalized == "" {
		return []ranking.Suggestion{}
	}

	rankTree := ranking.NewBST()
	seen := make(map[string]struct{}, len(dictWords))

	for _, dictWord := range dictWords {
		candidate := strings.ToLower(strings.TrimSpace(dictWord))
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}

		dist := distance.Levenshtein(normalized, candidate)
		if dist > 2 {
			continue
		}

		rankTree.Insert(ranking.Suggestion{
			Word:      candidate,
			Distance:  dist,
			Frequency: dictTrie.Frequency(candidate),
		})
	}

	return rankTree.Sorted(limit)
}
