package spellcheck

import (
	"testing"

	"spellchecker/trie"
)

func TestTokenize(t *testing.T) {
	text := "Hello, WORLD! કેમ છો? 123"
	got := Tokenize(text)
	want := []string{"hello", "world", "કેમ", "છો", "123"}

	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected token at %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestMisspelledWords(t *testing.T) {
	dictTrie := trie.NewTrie()
	dictTrie.Insert("hello", 3)
	dictTrie.Insert("world", 2)

	got := MisspelledWords("hello wrld world wrld", dictTrie)
	if len(got) != 1 || got[0] != "wrld" {
		t.Fatalf("expected [wrld], got %+v", got)
	}
}

func TestAutoCompleteWrapper(t *testing.T) {
	dictTrie := trie.NewTrie()
	dictTrie.Insert("program", 10)
	dictTrie.Insert("progress", 7)
	dictTrie.Insert("project", 3)

	got := AutoComplete(" ProG ", dictTrie, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(got))
	}
	if got[0].Word != "program" || got[1].Word != "progress" {
		t.Fatalf("unexpected autocomplete order: %+v", got)
	}
}

func TestDidYouMean(t *testing.T) {
	dictTrie := trie.NewTrie()
	dictTrie.Insert("program", 10)
	dictTrie.Insert("progress", 7)
	dictTrie.Insert("project", 3)

	dictWords := []string{"program", "progress", "project"}
	got := DidYouMean("progrm", dictWords, dictTrie, 5)

	if len(got) == 0 {
		t.Fatal("expected suggestions, got none")
	}
	if got[0].Word != "program" || got[0].Distance != 1 {
		t.Fatalf("expected top suggestion program(distance=1), got %+v", got[0])
	}
}

