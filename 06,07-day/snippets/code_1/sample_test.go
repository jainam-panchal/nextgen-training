package main

import "testing"

func failWithoutHelper(t *testing.T, got int, want int) {
	if got != want {
		t.Fatalf("without helper: got %d, want %d", got, want)
	}
}

func failWithHelper(t *testing.T, got int, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("with helper: got %d, want %d", got, want)
	}
}

func TestWithoutHelper(t *testing.T) {
	failWithoutHelper(t, 1, 2)
}

func TestWithHelper(t *testing.T) {
	failWithHelper(t, 1, 2)
}
