package router

import (
	"testing"
	"time"

	"jainamp-panchal/nextgen-training/packet-router/internal/classifier"
	"jainamp-panchal/nextgen-training/packet-router/internal/packet"
	"jainamp-panchal/nextgen-training/packet-router/internal/queue"
)

func TestRouterDequeuesHighestPriorityFirst(t *testing.T) {
	r := NewRouter(queue.NewPacketLinkedListQueue, classifier.DefaultClassifier{})
	r.SetProcessingDelay(10 * time.Millisecond)
	low := &packet.Packet{
		ID:            1,
		TTL:           10,
		Priority:      5,
		Protocol:      packet.ProtocolUDP,
		SourceIP:      "10.0.0.1",
		DestinationIP: "10.0.0.2",
		Timestamp:     time.Now(),
	}

	high := &packet.Packet{
		ID:            2,
		TTL:           10,
		Priority:      5,
		Protocol:      packet.ProtocolICMP,
		ICMPType:      packet.ICMPPing,
		SourceIP:      "10.0.0.3",
		DestinationIP: "10.0.0.4",
		Timestamp:     time.Now(),
	}

	r.Enqueue(low)
	r.Enqueue(high)

	got, ok := r.Dequeue()
	if !ok {
		t.Fatal("expected packet, got none")
	}

	if got.ID != high.ID {
		t.Fatalf("expected packet %d first, got %d", high.ID, got.ID)
	}
}

func TestRouterDropsExpiredPacketOnEnqueue(t *testing.T) {
	r := NewRouter(queue.NewPacketLinkedListQueue, classifier.DefaultClassifier{})
	r.SetProcessingDelay(10 * time.Millisecond)
	expired := &packet.Packet{
		ID:       10,
		TTL:      0,
		Priority: 1,
		Protocol: packet.ProtocolTCP,
	}

	ok := r.Enqueue(expired)
	if ok {
		t.Fatal("expected expired packet to be rejected")
	}

	if r.Len() != 0 {
		t.Fatalf("expected router length 0, got %d", r.Len())
	}
}

func TestRouterStatus(t *testing.T) {
	r := NewRouter(queue.NewPacketLinkedListQueue, classifier.DefaultClassifier{})
	r.SetProcessingDelay(10 * time.Millisecond)
	r.Enqueue(&packet.Packet{
		ID:       1,
		TTL:      10,
		Priority: 3,
		Protocol: packet.ProtocolUDP,
	})

	r.Enqueue(&packet.Packet{
		ID:       2,
		TTL:      10,
		Priority: 5,
		Protocol: packet.ProtocolTCP,
	})

	status := r.Status()

	if len(status) != packet.PriorityLevels {
		t.Fatalf("expected %d priority statuses, got %d", packet.PriorityLevels, len(status))
	}

	if status[2].Priority != 3 || status[2].Length != 1 {
		t.Fatalf("expected priority 3 length 1, got priority %d length %d", status[2].Priority, status[2].Length)
	}

	if status[4].Priority != 5 || status[4].Length != 1 {
		t.Fatalf("expected priority 5 length 1, got priority %d length %d", status[4].Priority, status[4].Length)
	}
}

func TestRouterProcessNextDecrementsTTL(t *testing.T) {
	r := NewRouter(queue.NewPacketLinkedListQueue, classifier.DefaultClassifier{})
	r.SetProcessingDelay(10 * time.Millisecond)
	r.Enqueue(&packet.Packet{
		ID:       1,
		TTL:      5,
		Priority: 1,
		Protocol: packet.ProtocolUDP,
	})

	got, ok := r.ProcessNext()
	if !ok {
		t.Fatal("expected packet to be processed")
	}

	if got == nil {
		t.Fatal("expected processed packet, got nil")
	}

	if got.TTL != 4 {
		t.Fatalf("expected TTL 4, got %d", got.TTL)
	}
}

func TestRouterProcessNextDropsPacketWhenTTLExpires(t *testing.T) {
	r := NewRouter(queue.NewPacketLinkedListQueue, classifier.DefaultClassifier{})
	r.SetProcessingDelay(10 * time.Millisecond)
	r.Enqueue(&packet.Packet{
		ID:       1,
		TTL:      1,
		Priority: 1,
		Protocol: packet.ProtocolUDP,
	})

	got, ok := r.ProcessNext()
	if !ok {
		t.Fatal("expected packet to be consumed")
	}

	if got != nil {
		t.Fatalf("expected packet to be dropped after TTL expiry, got packet ID %d", got.ID)
	}

	if r.Len() != 0 {
		t.Fatalf("expected router length 0, got %d", r.Len())
	}
}
