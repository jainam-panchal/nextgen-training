package classifier

import "jainamp-panchal/nextgen-training/packet-router/internal/packet"

type PacketClassifier interface {
	Classify(p *packet.Packet) int
}