// Package dnsrecord
package dnsrecord

import (
	"time"
)

type DNSRecord struct {
	Domain    string        `json:"domain"`
	IP        string        `json:"ip"`
	HitCount  int           `json:"hit_count"`
	TTL       time.Duration `json:"ttl"`
	CreatedAt time.Time     `json:"created_at"`
}

func (r DNSRecord) IsExpired(now time.Time) bool {
	return now.Sub(r.CreatedAt) >= r.TTL
}
