package resolver

import (
	"context"
	cachestore "dns-cache-resolver/internal/cache_store"
	dnsrecord "dns-cache-resolver/internal/dns_record"
	"dns-cache-resolver/internal/upstream"
	"fmt"
	"strings"
	"time"
)

const defaultTTL = 30 * time.Second

type Resolver struct {
	Store    cachestore.CacheStore[dnsrecord.DNSRecord]
	Upstream upstream.UpstreamResolver

	hits   int64
	misses int64
}

type CacheStats struct {
	Hits           int64
	Misses         int64
	TotalEntries   int
	HitRate        float64
	MissRate       float64
	MemoryEstimate int64
}

func (r *Resolver) Stats() CacheStats {
	totalLookups := r.hits + r.misses

	var hitRate, missRate float64
	if totalLookups > 0 {
		hitRate = float64(r.hits) / float64(totalLookups)
		missRate = float64(r.misses) / float64(totalLookups)
	}

	entries := r.Store.Len()

	const approxBytesPerEntry = 160
	mem := int64(entries * approxBytesPerEntry)

	return CacheStats{
		Hits:           r.hits,
		Misses:         r.misses,
		TotalEntries:   entries,
		HitRate:        hitRate,
		MissRate:       missRate,
		MemoryEstimate: mem,
	}
}

func NewResolver(store cachestore.CacheStore[dnsrecord.DNSRecord], resolver upstream.UpstreamResolver) *Resolver {
	return &Resolver{
		Store:    store,
		Upstream: resolver,
	}
}

func (r *Resolver) Resolve(domain string) (string, error) {
	now := time.Now()

	record, ok := r.lookupValidRecord(domain, now)
	if !ok {
		for _, wildcard := range wildcardCandidates(domain) {
			record, ok = r.lookupValidRecord(wildcard, now)
			if ok {
				break
			}
		}
	}

	if ok {
		r.hits++
		record.HitCount++
		r.Store.Set(record.Domain, record)
		return record.IP, nil
	}

	r.misses++

	ipAddress, err := r.Upstream.Resolve(domain)
	if err != nil {
		return "", err
	}

	newRecord := dnsrecord.DNSRecord{
		Domain:    domain,
		IP:        ipAddress,
		HitCount:  0,
		TTL:       defaultTTL,
		CreatedAt: now,
	}

	r.Store.Set(newRecord.Domain, newRecord)
	return ipAddress, nil
}

func wildcardCandidates(domain string) []string {
	parts := strings.Split(domain, ".")
	if len(parts) < 3 {
		return nil
	}

	candidates := make([]string, 0, len(parts)-2)
	for i := 1; i <= len(parts)-2; i++ {
		candidates = append(candidates, "*."+strings.Join(parts[i:], "."))
	}

	return candidates
}

func (r *Resolver) lookupValidRecord(domain string, now time.Time) (dnsrecord.DNSRecord, bool) {
	record, ok := r.Store.Get(domain)
	if !ok {
		var zero dnsrecord.DNSRecord
		return zero, false
	}

	if record.IsExpired(now) {
		r.Store.Delete(domain)
		var zero dnsrecord.DNSRecord
		return zero, false
	}

	return record, true
}

func (r *Resolver) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.cleanupExpired(time.Now())
			}
		}
	}()
}

func (r *Resolver) cleanupExpired(now time.Time) {
	for _, key := range r.Store.Keys() {
		record, ok := r.Store.Get(key)
		if !ok {
			continue
		}
		if record.IsExpired(now) {
			r.Store.Delete(key)
		}
	}
}

func (r *Resolver) AddRecord(domain, ip string, ttl time.Duration) error {
	if strings.TrimSpace(domain) == "" {
		return fmt.Errorf("invalid domain")
	}
	if strings.TrimSpace(ip) == "" {
		return fmt.Errorf("invalid ip")
	}
	if ttl <= 0 {
		return fmt.Errorf("ttl must be > 0")
	}

	record := dnsrecord.DNSRecord{
		Domain:    domain,
		IP:        ip,
		TTL:       ttl,
		CreatedAt: time.Now(),
		HitCount:  0,
	}

	r.Store.Set(domain, record)
	return nil
}
