package distance

import "testing"

func TestLevenshteinBasicCases(t *testing.T) {
	tests := []struct {
		name   string
		first  string
		second string
		want   int
	}{
		{name: "same", first: "program", second: "program", want: 0},
		{name: "insert", first: "progrm", second: "program", want: 1},
		{name: "delete", first: "program", second: "progrm", want: 1},
		{name: "replace", first: "program", second: "progrum", want: 1},
		{name: "kitten-sitting", first: "kitten", second: "sitting", want: 3},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := Levenshtein(test.first, test.second)
			if got != test.want {
				t.Fatalf("Levenshtein(%q, %q)=%d, want %d", test.first, test.second, got, test.want)
			}
		})
	}
}

func TestLevenshteinUnicode(t *testing.T) {
	got := Levenshtein("કેમ", "કૈમ")
	if got != 1 {
		t.Fatalf("expected unicode distance 1, got %d", got)
	}
}

