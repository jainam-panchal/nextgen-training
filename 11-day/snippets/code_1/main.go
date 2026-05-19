package main

import (
	"fmt"
	"time"
)

type Message struct {
	From    string
	To      string
	Message string
}

func producer(in chan<- Message) {
	defer close(in)
	for i := 0; i < 10; i++ {
		in <- Message{
			From:    fmt.Sprintf("From %d", i),
			To:      fmt.Sprintf("To %d", i),
			Message: fmt.Sprintf("Message %d", i),
		}
	}

	fmt.Println("producer done")
}

func worker(in <-chan Message, out chan<- Message) {
	defer close(out) // worker closes output after input is drained

	for msg := range in {
		fmt.Println("worker processing:", msg.Message)
		time.Sleep(500 * time.Millisecond)

		msg.Message = msg.Message + " processed"
		out <- msg
	}
}

func printer(in <-chan Message, done chan<- struct{}) {
	defer close(done)

	for msg := range in {
		fmt.Println("printer received:", msg)
	}

	fmt.Println("pipeline fully drained")
}

func main() {
	in := make(chan Message, 10)
	out := make(chan Message, 10)
	done := make(chan struct{})

	go producer(in)
	go worker(in, out)
	go printer(out, done)

	<-done
	fmt.Println("main exiting cleanly")
}

//package main
//
//import (
//	"fmt"
//	"time"
//)
//
//type Message struct {
//	From    string
//	To      string
//	Message string
//}
//
//func producer(out chan<- Message) {
//	defer close(out)
//
//	for i := 1; i <= 5; i++ {
//		out <- Message{
//			From:    "Jainam",
//			To:      "Server",
//			Message: fmt.Sprintf("message-%d", i),
//		}
//
//		time.Sleep(300 * time.Millisecond)
//	}
//}
//
//func worker(in <-chan Message, out chan<- Message) {
//	defer close(out) // worker closes output after input is drained
//
//	for msg := range in {
//		fmt.Println("worker processing:", msg.Message)
//		time.Sleep(500 * time.Millisecond)
//
//		msg.Message = msg.Message + " processed"
//		out <- msg
//	}
//}
//
//func printer(in <-chan Message, done chan<- struct{}) {
//	defer close(done)
//
//	for msg := range in {
//		fmt.Println("printer received:", msg)
//	}
//
//	fmt.Println("pipeline fully drained")
//}
//
//func main() {
//	inputChan := make(chan Message, 3)
//	outputChan := make(chan Message, 3)
//	done := make(chan struct{})
//
//	go producer(inputChan)
//	go worker(inputChan, outputChan)
//	go printer(outputChan, done)
//
//	<-done
//	fmt.Println("main exiting cleanly")
//}
