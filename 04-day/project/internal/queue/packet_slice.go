package queue

import "jainamp-panchal/nextgen-training/packet-router/internal/packet"

func NewPacketSliceQueue() PacketQueue {
	return NewSliceQueue[*packet.Packet]()
}

var _ PacketQueue = (*SliceQueue[*packet.Packet])(nil)
