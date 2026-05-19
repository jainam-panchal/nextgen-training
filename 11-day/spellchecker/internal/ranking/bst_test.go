package ranking

import "testing"

func TestBSTSortedRankingOrder(t *testing.T) {
	tree := NewBST()

	tree.Insert(Suggestion{Word: "progress", Distance: 2, Frequency: 9})
	tree.Insert(Suggestion{Word: "program", Distance: 1, Frequency: 3})
	tree.Insert(Suggestion{Word: "programmer", Distance: 1, Frequency: 7})
	tree.Insert(Suggestion{Word: "prologue", Distance: 1, Frequency: 7})
	tree.Insert(Suggestion{Word: "project", Distance: 3, Frequency: 100})

	got := tree.Sorted(10)
	want := []Suggestion{
		{Word: "programmer", Distance: 1, Frequency: 7},
		{Word: "prologue", Distance: 1, Frequency: 7},
		{Word: "program", Distance: 1, Frequency: 3},
		{Word: "progress", Distance: 2, Frequency: 9},
		{Word: "project", Distance: 3, Frequency: 100},
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d suggestions, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected suggestion at %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestBSTSortedLimit(t *testing.T) {
	tree := NewBST()
	tree.Insert(Suggestion{Word: "c", Distance: 2, Frequency: 1})
	tree.Insert(Suggestion{Word: "a", Distance: 1, Frequency: 1})
	tree.Insert(Suggestion{Word: "b", Distance: 1, Frequency: 1})

	got := tree.Sorted(2)
	if len(got) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(got))
	}

	if got[0].Word != "a" || got[1].Word != "b" {
		t.Fatalf("unexpected order with limit: got %+v", got)
	}
}
