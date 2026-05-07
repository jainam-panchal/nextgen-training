package main

import (
	"fmt"
	"time"

	"jainamp-panchal/nextgen-training/packet-router/internal/classifier"
	"jainamp-panchal/nextgen-training/packet-router/internal/packet"
	"jainamp-panchal/nextgen-training/packet-router/internal/queue"
	"jainamp-panchal/nextgen-training/packet-router/internal/router"
)

func main() {
	r := router.NewRouter(
		queue.NewPacketRingBufferQueue,
		classifier.DefaultClassifier{},
	)

	r.SetProcessingDelay(10 * time.Millisecond)

	packets := []*packet.Packet{
		{
			ID:            1,
			TTL:           10,
			Priority:      5,
			Protocol:      packet.ProtocolUDP,
			SourceIP:      "10.0.0.1",
			DestinationIP: "10.0.0.2",
			Timestamp:     time.Now(),
			Payload:       []byte("normal udp packet"),
		},
		{
			ID:            2,
			TTL:           10,
			Priority:      5,
			Protocol:      packet.ProtocolICMP,
			ICMPType:      packet.ICMPPing,
			SourceIP:      "10.0.0.3",
			DestinationIP: "10.0.0.4",
			Timestamp:     time.Now(),
			Payload:       []byte("ping"),
		},
		{
			ID:            3,
			TTL:           10,
			Priority:      5,
			Protocol:      packet.ProtocolTCP,
			TCPFlags:      packet.TCPFlags{SYN: true},
			SourceIP:      "10.0.0.5",
			DestinationIP: "10.0.0.6",
			Timestamp:     time.Now(),
			Payload:       []byte("tcp syn"),
		},
		{
			ID:            4,
			TTL:           0,
			Priority:      1,
			Protocol:      packet.ProtocolTCP,
			SourceIP:      "10.0.0.7",
			DestinationIP: "10.0.0.8",
			Timestamp:     time.Now(),
			Payload:       []byte("expired packet"),
		},
	}

	for _, p := range packets {
		ok := r.Enqueue(p)
		if !ok {
			fmt.Printf("packet ID %d was rejected\n", p.ID)
		}
	}

	fmt.Println("queue status after enqueue:")
	printStatus(r.Status())

	fmt.Println("\nprocessing packets:")
	for {
		p, ok := r.ProcessNext()
		if !ok {
			break
		}

		if p == nil {
			fmt.Println("packet dropped during processing")
			continue
		}

		fmt.Println("processed", p)
	}

	fmt.Println("\nqueue status after processing:")
	printStatus(r.Status())

	fmt.Println("\ndropped packets:")
	for _, d := range r.Drops() {
		fmt.Printf("packet ID=%d reason=%s time=%s\n",
			d.PacketID,
			d.Reason,
			d.Timestamp.Format(time.RFC3339),
		)
	}
}

func printStatus(statuses []router.QueueStatus) {
	for _, s := range statuses {
		fmt.Printf("priority %d: %d packets\n", s.Priority, s.Length)
	}
}
