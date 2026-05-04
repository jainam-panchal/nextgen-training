package main

import "fmt"

// Array passed by value and Slices are passed by ref
func do(a [3]int, b []int) []int {
	// a = b // Cannot assign a slice to an array

	a[0] = 4 // w array is unchanged
	b[0] = 9 // x slice is changed

	c := make([]int, 5) // slice of 5 zeros

	c[4] = 66

	copy(c, b) // first size(b) values will be copied from b to c

	return c // C => [ 9 , 0 , 0 , 0 , 66 ]
}

func main() {
	w := [...]int{1, 2, 3} // array literal with length determined by the number of elements

	x := []int{0, 0, 0} // slice literal

	y := do(w, x)

	fmt.Println(w, x, y)

	// --------------------------------------
	fmt.Println("---------------------------")

	slice1 := []int{1, 2, 3}

	slice2 := make([]byte, len(slice1), (cap(slice1)+1)*2)

	fmt.Println("slice1", len(slice1), len(slice1))
	fmt.Println("slice2", len(slice2), len(slice2))
}
