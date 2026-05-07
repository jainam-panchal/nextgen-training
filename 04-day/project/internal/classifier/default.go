// Package classifier
package classifier

import "jainamp-panchal/nextgen-training/packet-router/internal/packet"

type DefaultClassifier struct{}

func (DefaultClassifier) Classify(p *packet.Packet) int {
	if p == nil {
		return packet.LowestPriority
	}

	if p.Protocol == packet.ProtocolICMP && p.ICMPType == packet.ICMPPing {
		return packet.HighestPriority
	}

	if p.Protocol == packet.ProtocolTCP && p.TCPFlags.SYN && !p.TCPFlags.ACK {
		return 2
	}

	return packet.NormalizePriority(p.Priority)
}
