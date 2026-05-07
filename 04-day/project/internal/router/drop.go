package router

import (
	"time"

	"jainamp-panchal/nextgen-training/packet-router/internal/packet"
)

type DropReason string

const (
	DropReasonTTLExpired DropReason = "ttl_expired"
	DropReasonNilPacket  DropReason = "nil_packet"
)

type DropEvent struct {
	PacketID  int
	Reason    DropReason
	Timestamp time.Time
	Packet    *packet.Packet
}

func (r *Router) logDrop(p *packet.Packet, reason DropReason) {
	packetID := 0
	if p != nil {
		packetID = p.ID
	}

	r.drops = append(r.drops, DropEvent{
		PacketID:  packetID,
		Reason:    reason,
		Timestamp: time.Now(),
		Packet:    p,
	})
}

func (r *Router) Drops() []DropEvent {
	drops := make([]DropEvent, len(r.drops))
	copy(drops, r.drops)
	return drops
}
