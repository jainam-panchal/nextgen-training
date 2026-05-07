package queue

import "jainamp-panchal/nextgen-training/packet-router/internal/packet"

func NewPacketRingBufferQueue() PacketQueue {
	return NewRingBufferQueue[*packet.Packet]()
}

var _ PacketQueue = (*RingBufferQueue[*packet.Packet])(nil)
