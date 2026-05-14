package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Printf("goroutines at start: %d\n", runtime.NumGoroutine())

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		i := i  // capture loop variable (required pre Go 1.22)
		go func() {
			defer wg.Done()
			fmt.Printf("goroutine %d starting\n", i)
			time.Sleep(time.Duration(i) * 100 * time.Millisecond)
			fmt.Printf("goroutine %d done\n", i)
		}()
	}

	fmt.Printf("goroutines after launch: %d\n", runtime.NumGoroutine())
	wg.Wait()
	fmt.Printf("goroutines after wait: %d\n", runtime.NumGoroutine())
}