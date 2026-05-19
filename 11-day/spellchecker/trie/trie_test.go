package trie

import "testing"

func TestInsertAndSearchSingleWord(t *testing.T) {
	trie := NewTrie()

	trie.Insert("program", 1)

	if !trie.Search("program") {
		t.Fatal("expected word to exist")
	}
}
func TestSearchReturnsFalseForMissingWord(t *testing.T) {
	trie := NewTrie()

	trie.Insert("program", 1)

	if trie.Search("progress") {
		t.Fatal("expected missing word to return false")
	}
}

func TestSearchReturnsFalseForPrefixOnly(t *testing.T) {
	trie := NewTrie()

	trie.Insert("program", 1)

	if trie.Search("pro") {
		t.Fatal("expected prefix to not be treated as word")
	}
}

func TestInsertSupportsUnicode(t *testing.T) {
	trie := NewTrie()

	trie.Insert("કેમ", 1)

	if !trie.Search("કેમ") {
		t.Fatal("expected unicode word to exist")
	}
}

func TestStartsWithReturnsTrueForExistingPrefix(t *testing.T) {
	trie := NewTrie()

	trie.Insert("program", 1)

	if !trie.StartsWith("pro") {
		t.Fatal("expected prefix to exist")
	}
}

func TestStartsWithReturnsFalseForMissingPrefix(t *testing.T) {
	trie := NewTrie()

	trie.Insert("program", 1)

	if trie.StartsWith("xyz") {
		t.Fatal("expected missing prefix to return false")
	}
}

func TestInsertAccumulatesFrequency(t *testing.T) {
	trie := NewTrie()

	trie.Insert("program", 2)
	trie.Insert("program", 3)

	if got := trie.Frequency("program"); got != 5 {
		t.Fatalf("expected frequency 5, got %d", got)
	}
}

func TestFrequencyReturnsZeroForPrefix(t *testing.T) {
	trie := NewTrie()

	trie.Insert("program", 2)

	if got := trie.Frequency("pro"); got != 0 {
		t.Fatalf("expected frequency 0 for prefix, got %d", got)
	}
}
func TestAutoCompleteReturnsSuggestionsSortedByFrequency(t *testing.T) {
	trie := NewTrie()

	trie.Insert("program", 10)
	trie.Insert("progress", 7)
	trie.Insert("programming", 15)
	trie.Insert("project", 3)

	got := trie.AutoComplete("prog", 5)

	want := []string{"programming", "program", "progress"}

	if len(got) != len(want) {
		t.Fatalf("expected %d suggestions, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i].Word != want[i] {
			t.Fatalf("expected %q at index %d, got %q", want[i], i, got[i].Word)
		}
	}
}

func TestDeleteExistingWord(t *testing.T) {
	trie := NewTrie()
	trie.Insert("program", 3)

	if !trie.Delete("program") {
		t.Fatal("expected delete to succeed for existing word")
	}

	if trie.Search("program") {
		t.Fatal("expected deleted word to be missing")
	}
}

func TestDeleteMissingWordReturnsFalse(t *testing.T) {
	trie := NewTrie()
	trie.Insert("program", 1)

	if trie.Delete("progress") {
		t.Fatal("expected delete to return false for missing word")
	}
}

func TestDeleteKeepsOverlappingWord(t *testing.T) {
	trie := NewTrie()
	trie.Insert("program", 1)
	trie.Insert("programmer", 2)

	if !trie.Delete("program") {
		t.Fatal("expected delete of 'program' to succeed")
	}

	if trie.Search("program") {
		t.Fatal("expected 'program' to be deleted")
	}

	if !trie.Search("programmer") {
		t.Fatal("expected overlapping word 'programmer' to remain")
	}
}

func TestDeleteSupportsUnicode(t *testing.T) {
	trie := NewTrie()
	trie.Insert("કેમ", 2)

	if !trie.Delete("કેમ") {
		t.Fatal("expected unicode word delete to succeed")
	}

	if trie.Search("કેમ") {
		t.Fatal("expected unicode word to be deleted")
	}
}

func TestDeleteResetsFrequency(t *testing.T) {
	trie := NewTrie()
	trie.Insert("program", 5)

	if !trie.Delete("program") {
		t.Fatal("expected delete to succeed")
	}

	if got := trie.Frequency("program"); got != 0 {
		t.Fatalf("expected frequency to reset to 0, got %d", got)
	}
}
