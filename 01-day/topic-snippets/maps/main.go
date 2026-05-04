package main

import "fmt"

/*

Maps are dictionaries
- Can read from a nil map, but insertion will panic
- Passed by reference. No copying but updating is okay
- IMP.. Type used for the key must have == and != defined (not slices , maps , or funcs)

- A map cna't be compared to another one. Can be compared only with as a special case
*/

func main() {
	var m map[string]int      // nil , no storage
	p := make(map[string]int) // non-nil but empty

	a := p["the"] // returns 0

	b := m["the"] // same thing

	// m["and"] = 1 // -> Will crash as m is empty

	m = p      // m now points to p's memory
	m["and"]++ // will work as it points to P

	c := p["and"] // returns 1

	fmt.Println(a, b, c)

	// --------------------------------------
	fmt.Println("---------------------------")

	mp1 := map[string]int{
		"the": 2,
		"and": 1,
		"or":  3,
	}

	var mp2 map[string]int

	// cmp1 := mp1 == mp2 // Syntax error -> invalid operation: mp1 == mp2 (map can only be compared to nil)
	cmp2 := mp2 == nil
	cmp3 := len(mp1)
	// cmp4 := cap(mp1)  // Type Mismatch -> invalid argument: mp1 (variable of type map[string]int) for built-in cap

	fmt.Println(cmp2, cmp3)

	// --------------------------------------
	fmt.Println("---------------------------")

	// Two way lookup

	mp3 := map[string]int{} // non-nil but empty

	value1 := mp3["the"]
	value2, ok := mp3["the"] // 0 , false

	fmt.Println(value1, value2, ok)

	mp3["randomValue"]++
	value3 := mp3["randomValue"]
	value4, ok := mp3["randomValue"] // 1, true

	fmt.Println(value3, value4, ok)
}
