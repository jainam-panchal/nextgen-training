package router

import (
	"fmt"
	"testing"
	"time"

	"jainamp-panchal/nextgen-training/packet-router/internal/classifier"
	"jainamp-panchal/nextgen-training/packet-router/internal/packet"
	"jainamp-panchal/nextgen-training/packet-router/internal/queue"
)

func benchmarkRouter(b *testing.B, n int, newQueue queue.NewQueueFunc) {
	packets := generatePackets(n)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := NewRouter(newQueue, classifier.DefaultClassifier{})

		for _, p := range packets {
			cp := *p
			r.Enqueue(&cp)
		}

		for {
			_, ok := r.ProcessNext()
			if !ok {
				break
			}
		}
	}
}

func generatePackets(n int) []*packet.Packet {
	packets := make([]*packet.Packet, 0, n)

	for i := 0; i < n; i++ {
		p := &packet.Packet{
			ID:            i + 1,
			TTL:           64,
			Priority:      (i % packet.PriorityLevels) + 1,
			SourceIP:      "10.0.0.1",
			DestinationIP: "10.0.0.2",
			Timestamp:     time.Now(),
			Payload:       []byte("payload"),
		}

		switch i % 3 {
		case 0:
			p.Protocol = packet.ProtocolICMP
			p.ICMPType = packet.ICMPPing
		case 1:
			p.Protocol = packet.ProtocolTCP
			p.TCPFlags = packet.TCPFlags{SYN: true}
		default:
			p.Protocol = packet.ProtocolUDP
		}

		packets = append(packets, p)
	}

	return packets
}

func BenchmarkRouterLinkedList(b *testing.B) {
	for _, n := range []int{
		100,
		1_000,
		10_000,
		100_000,
		500_000,
		1_000_000,
	} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			benchmarkRouter(b, n, queue.NewPacketLinkedListQueue)
		})
	}
}

func BenchmarkRouterSlice(b *testing.B) {
	for _, n := range []int{
		100,
		1_000,
		10_000,
		100_000,
		500_000,
		1_000_000,
	} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			benchmarkRouter(b, n, queue.NewPacketSliceQueue)
		})
	}
}

func BenchmarkRouterRingBuffer(b *testing.B) {
	for _, n := range []int{
		100,
		1_000,
		10_000,
		100_000,
		500_000,
		1_000_000,
	} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			benchmarkRouter(b, n, queue.NewPacketRingBufferQueue)
		})
	}
}
