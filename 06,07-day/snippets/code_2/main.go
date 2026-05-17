package main

import (
	"fmt"
	"sync"
)

func main() {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	ready := false
	workerWaiting := make(chan struct{})

	go func() {
		mu.Lock()

		close(workerWaiting)

		for !ready {
			fmt.Println("worker waiting")
			cond.Wait()
		}

		fmt.Println("worker: ready")
		mu.Unlock()
	}()
	package main

	<-workerWaiting

	mu.Lock()
	ready = true
	cond.Signal()
	mu.Unlock()
}
