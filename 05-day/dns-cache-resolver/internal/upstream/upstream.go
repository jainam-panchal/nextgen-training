// Package upstream
package upstream

import (
	"fmt"
	"hash/fnv"
	"net"
)

type UpstreamResolver interface {
	Resolve(string) (string, error)
}

type DummyUpstreamResolver struct {
}

func NewDummyUpstreamResolver() *DummyUpstreamResolver {
	return &DummyUpstreamResolver{}
}

func (*DummyUpstreamResolver) Resolve(domain string) (string, error) {
	if domain == "" {
		return "", fmt.Errorf("invalid domain")
	}

	// 32-bit FNV-1a hasher
	h := fnv.New32a()
	_, _ = h.Write([]byte(domain))
	n := h.Sum32()

	ip := net.IPv4(
		203,
		0,
		113,
		byte((n%250)+1),
	)

	return ip.String(), nil
}
