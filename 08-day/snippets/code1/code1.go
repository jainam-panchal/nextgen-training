package main

import "fmt"

func panicAndRecover() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", r)
		}
	}()

	panic("Boom!")

	fmt.Println("Hello World")
}

func main() {
	panicAndRecover()
}
