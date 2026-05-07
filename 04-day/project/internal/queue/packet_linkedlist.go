package queue

import "jainamp-panchal/nextgen-training/packet-router/internal/packet"

func NewPacketLinkedListQueue() PacketQueue {
	return NewLinkedListQueue[*packet.Packet]()
}
