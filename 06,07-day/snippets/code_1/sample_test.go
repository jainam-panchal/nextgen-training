package main

import "testing"

func assertEqual(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

func TestAdd(t *testing.T) {
	result := 2 + 2

	assertEqual(t, result, 5)
}
