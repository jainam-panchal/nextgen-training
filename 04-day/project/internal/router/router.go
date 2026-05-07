// Package router
package router

import (
	"time"

	"jainamp-panchal/nextgen-training/packet-router/internal/classifier"
	"jainamp-panchal/nextgen-training/packet-router/internal/packet"
	"jainamp-panchal/nextgen-training/packet-router/internal/queue"
)

type Router struct {
	queues          [packet.PriorityLevels]queue.PacketQueue
	classifier      classifier.PacketClassifier
	processingDelay time.Duration
	drops           []DropEvent
}

func NewRouter(
	newQueue queue.NewQueueFunc,
	packetClassifer classifier.PacketClassifier,
) *Router {
	r := &Router{
		classifier: packetClassifer,
	}

	for i := 0; i < packet.PriorityLevels; i++ {
		r.queues[i] = newQueue()
	}

	return r
}

func (r *Router) Enqueue(p *packet.Packet) bool {
	if p == nil {
		r.logDrop(nil, DropReasonNilPacket)
		return false
	}

	if p.TTL <= 0 {
		r.logDrop(p, DropReasonTTLExpired)
		return false
	}

	priority := r.classifier.Classify(p)
	priority = packet.NormalizePriority(priority)

	p.Priority = priority // Mutate the updated priority

	r.queues[priority-1].Enqueue(p)

	return true
}

func (r *Router) Dequeue() (*packet.Packet, bool) {
	for i := 0; i < packet.PriorityLevels; i++ {
		p, ok := r.queues[i].Dequeue()
		if ok {
			return p, true
		}
	}

	return nil, false
}

func (r *Router) Len() int {
	total := 0

	for i := 0; i < packet.PriorityLevels; i++ {
		total += r.queues[i].Len()
	}

	return total
}

type QueueStatus struct {
	Priority int
	Length   int
}

func (r *Router) Status() []QueueStatus {
	statuses := make([]QueueStatus, 0, packet.PriorityLevels)

	for i := 0; i < packet.PriorityLevels; i++ {
		statuses = append(statuses, QueueStatus{
			Priority: i + 1,
			Length:   r.queues[i].Len(),
		})
	}

	return statuses
}

func (r *Router) ProcessNext() (*packet.Packet, bool) {
	p, ok := r.Dequeue()
	if !ok {
		return nil, false
	}

	if r.processingDelay > 0 {
		time.Sleep(r.processingDelay)
	}

	p.TTL--

	if p.TTL <= 0 {
		r.logDrop(p, DropReasonTTLExpired)
		return nil, true
	}

	return p, true
}

func (r *Router) SetProcessingDelay(delay time.Duration) {
	r.processingDelay = delay
}
