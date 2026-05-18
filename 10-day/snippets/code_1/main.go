package main

import (
	"fmt"
	"sync"
	"time"
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
			cond.Wait()
			fmt.Println("ready loop is exiting")
		}

		fmt.Println("worker: ready")
		mu.Unlock()
	}()

	<-workerWaiting

	mu.Lock()
	ready = true
	cond.Signal()
	fmt.Println("signalled now!")
	time.Sleep(5 * time.Second)
	mu.Unlock()
	fmt.Println("worker unlocked")
	time.Sleep(100 * time.Second)
}
