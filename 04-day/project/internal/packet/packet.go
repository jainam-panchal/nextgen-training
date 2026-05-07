// Package packet
package packet

import (
	"fmt"
	"time"
)

type (
	Protocol string
	ICMPType string
)

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolICMP Protocol = "ICMP"
)

const (
	ICMPPing      ICMPType = "PING"
	ICMPPingReply ICMPType = "PING_REPLY"
)

const (
	HighestPriority = 1
	LowestPriority  = 5
	PriorityLevels  = 5
)

type Packet struct {
	Payload []byte

	ID       int
	TTL      int
	Priority int

	SourceIP      string
	DestinationIP string

	Timestamp time.Time
	Protocol  Protocol

	TCPFlags TCPFlags
	ICMPType ICMPType
}

type TCPFlags struct {
	SYN bool
	ACK bool
	FIN bool
	RST bool
}

func NormalizePriority(priority int) int {
	if priority < HighestPriority {
		return HighestPriority
	}
	if priority > LowestPriority {
		return LowestPriority
	}
	return priority
}

func (p Packet) String() string {
	return fmt.Sprintf(
		"Packet{ID:%d Protocol:%s Source:%s Destination:%s Priority:%d TTL:%d PayloadBytes:%d}",
		p.ID,
		p.Protocol,
		p.SourceIP,
		p.DestinationIP,
		p.Priority,
		p.TTL,
		len(p.Payload),
	)
}
