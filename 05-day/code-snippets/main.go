package main

import (
	"fmt"
	"nextgen/lru"
)

func main() {
	cache := lru.NewLRUCache[string](2)

	cache.Put("a", "apple")
	cache.Put("b", "banana")

	cache.Put("c", "cherry")

	valA, okA := cache.Get("a") // Should be false (evicted)
	valB, okB := cache.Get("b") // Should be true
	valC, okC := cache.Get("c") // Should be true

	fmt.Printf("Key 'a': %v, %v\n", valA, okA)
	fmt.Printf("Key 'b': %v, %v\n", valB, okB)
	fmt.Printf("Key 'c': %v, %v\n", valC, okC)

	cache.Put("b", "blueberry")

	fmt.Printf("Final Cache Length: %d\n", cache.Len())
}
