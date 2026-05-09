package main

import (
	"context"
	"fmt"
	"time"

	cachestore "dns-cache-resolver/internal/cache_store"
	dnsrecord "dns-cache-resolver/internal/dns_record"
	"dns-cache-resolver/internal/resolver"
	stream "dns-cache-resolver/internal/upstream"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		store    cachestore.CacheStore[dnsrecord.DNSRecord]
		upstream stream.UpstreamResolver
	)

	// Custom Map Impl
	//store = cachestore.NewCustomMapStore[dnsrecord.DNSRecord]()

	// Defualt Map Impl
	store = cachestore.NewInbuiltMapStore[dnsrecord.DNSRecord]()

	upstream = stream.NewDummyUpstreamResolver()

	r := resolver.NewResolver(store, upstream)
	r.StartCleanup(ctx, 30*time.Second)

	ip, err := r.Resolve("example.com")
	if err != nil {
		fmt.Println("resolve error:", err)
		return
	}

	fmt.Println("example.com ->", ip)

	if err := r.AddRecord("*.example.com", "203.0.113.42", 45*time.Second); err != nil {
		fmt.Println("add record error:", err)
		return
	}

	ip, err = r.Resolve("sub.example.com")
	if err != nil {
		fmt.Println("resolve error:", err)
		return
	}
	fmt.Println("sub.example.com ->", ip)

	stats := r.Stats()

	fmt.Printf("stats: hits=%d misses=%d entries=%d hit_rate=%.2f miss_rate=%.2f mem_est=%d\n",
		stats.Hits, stats.Misses, stats.TotalEntries, stats.HitRate, stats.MissRate,
		stats.MemoryEstimate)

}
