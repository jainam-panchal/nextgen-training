package cachestore

import (
	"fmt"
	"testing"
)

// TestCustomMapStoreDeleteKeepsCollidedEntries ensures chained collision entries
// are preserved when deleting one key from the same bucket chain.
func TestCustomMapStoreDeleteKeepsCollidedEntries(t *testing.T) {
	store := NewCustomMapStore[int]()

	k1, k2, ok := findCollisionPair(store)
	if !ok {
		t.Fatal("could not find colliding keys")
	}

	store.Set(k1, 11)
	store.Set(k2, 22)

	if store.Len() != 2 {
		t.Fatalf("expected size 2, got %d", store.Len())
	}

	if !store.Delete(k1) {
		t.Fatalf("expected delete(%s) to succeed", k1)
	}

	if store.Len() != 1 {
		t.Fatalf("expected size 1 after delete, got %d", store.Len())
	}

	if _, ok := store.Get(k1); ok {
		t.Fatalf("expected key %q to be deleted", k1)
	}

	v, ok := store.Get(k2)
	if !ok {
		t.Fatalf("expected collided key %q to remain", k2)
	}
	if v != 22 {
		t.Fatalf("expected value 22 for key %q, got %d", k2, v)
	}
}

// TestCustomMapStoreKeysReturnsAllEntries verifies Keys returns every stored key.
func TestCustomMapStoreKeysReturnsAllEntries(t *testing.T) {
	store := NewCustomMapStore[int]()
	store.Set("a.example.com", 1)
	store.Set("b.example.com", 2)
	store.Set("c.example.com", 3)

	keys := store.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}

	seen := map[string]bool{}
	for _, k := range keys {
		seen[k] = true
	}

	for _, want := range []string{"a.example.com", "b.example.com", "c.example.com"} {
		if !seen[want] {
			t.Fatalf("expected key %q in Keys()", want)
		}
	}
}

// findCollisionPair brute-forces two distinct keys that map to the same bucket.
func findCollisionPair[T any](store *CustomMapStore[T]) (string, string, bool) {
	firstByBucket := map[int]string{}

	for i := 0; i < 200000; i++ {
		key := fmt.Sprintf("key-%d", i)
		b := store.bucketIndex(key)

		if first, exists := firstByBucket[b]; exists && first != key {
			return first, key, true
		}
		firstByBucket[b] = key
	}

	return "", "", false
}
